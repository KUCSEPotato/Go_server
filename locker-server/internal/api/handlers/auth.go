package handlers

import (
	"crypto/sha256"
	"encoding/base64"
	"log"
	"net"
	"regexp"
	"time"

	"github.com/KUCSEPotato/locker-server/internal/util"
	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5"         // ErrNoRows 등 에러 타입 사용
	"github.com/jackc/pgx/v5/pgxpool" // 커넥션 풀
	"github.com/redis/go-redis/v9"    // 의존성 주입 구조체에 포함 (여기선 직접 사용X)
)

// Deps: 핸들러들이 의존하는 리소스(DB, Redis 등)를 담는 구조체
// - 의존성 주입(DI) 방식으로 테스트/확장성이 좋아짐
type Deps struct {
	DB  *pgxpool.Pool // PostgreSQL 풀
	RDB *redis.Client // Redis 클라이언트(여기 파일에선 사용 안하지만 통일성 위해 포함)
}

// 요청, 응답 구조체 정의
type RegisterRequest struct {
	StudentID string `json:"student_id"`
	Name      string `json:"name"`
	Phone     string `json:"phone_number"`
}

type RegisterResponse struct {
	Message   string `json:"message"`
	StudentID string `json:"student_id"`
}

type LoginRequest struct {
	StudentID string `json:"student_id"`
	Name      string `json:"name"`
	Phone     string `json:"phone_number"`
}

type LoginResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

// token
type RefreshRequest struct {
	AccessToken  string `json:"access_token"` // 선택적: 블랙리스트용
	RefreshToken string `json:"refresh_token"`
}

type LogoutRequest struct {
	AccessToken  string `json:"access_token"`  // 선택적: 요청 body에서 받기
	RefreshToken string `json:"refresh_token"` // 반납할 refresh token
}

// 취약점 주의: hardcoded-credentials Embedding credentials in source code risks unauthorized access
type RefreshResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

type LogoutResponse struct {
	Message string `json:"message"`
}

// Register 핸들러: 새 사용자 등록 (학번/이름/전화번호)
// Register godoc
// @Summary      회원가입
// @Description  새 사용자를 등록합니다. 학번은 중복될 수 없습니다. 전화번호란에는 숫자만 허용합니다.
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        payload body RegisterRequest true "회원가입 정보"
// @Success      201 {object} RegisterResponse
// @Failure      400 {object} ErrorResponse
// @Failure      409 {object} ErrorResponse
// @Router       /auth/register [post]
func Register(d Deps) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// 1) 요청 바디(JSON) 파싱
		var req RegisterRequest
		if err := c.BodyParser(&req); err != nil {
			return fiber.ErrBadRequest
		}

		// 2) 기본 유효성 검사
		if req.StudentID == "" || req.Name == "" || req.Phone == "" {
			return fiber.NewError(fiber.StatusBadRequest, "missing required fields: student_id, name, phone_number")
		}

		// 3) 학번 형식 검증 (예: 숫자만, 길이 제한 등)
		if len(req.StudentID) != 10 {
			return fiber.NewError(fiber.StatusBadRequest, "invalid student_id format")
		}

		// 4) 전화번호 형식 간단 검증
		if len(req.Phone) < 10 || len(req.Phone) > 15 {
			return fiber.NewError(fiber.StatusBadRequest, "invalid phone_number format")
		}
		// 4-2) 전화번호에 숫자만 들어갔는지 검사
		matched, _ := regexp.MatchString(`^[0-9]+$`, req.Phone)
		if !matched {
			return fiber.NewError(fiber.StatusBadRequest, "only numeric characters are allowed in phone_number")
		}

		// 5) 이름 길이 검증
		if len(req.Name) < 2 || len(req.Name) > 20 {
			return fiber.NewError(fiber.StatusBadRequest, "invalid name length")
		}

		// 6) 중복 학번 체크
		var exists bool
		err := d.DB.QueryRow(c.Context(),
			`SELECT EXISTS(SELECT 1 FROM users WHERE student_id=$1)`,
			req.StudentID,
		).Scan(&exists)
		if err != nil {
			return fiber.ErrInternalServerError
		}
		if exists {
			return fiber.NewError(fiber.StatusConflict, "student_id already exists")
		}

		// 7) DB에 새 사용자 삽입
		_, err = d.DB.Exec(c.Context(),
			`INSERT INTO users (student_id, name, phone_number, created_at)
			 VALUES ($1, $2, $3, now())`,
			req.StudentID, req.Name, req.Phone,
		)
		if err != nil {
			// PostgreSQL 고유 제약 조건 위반 등의 경우도 고려
			log.Printf("Register: failed to insert user: %v", err)
			return fiber.ErrInternalServerError
		}

		// 8) 성공 응답
		return c.Status(fiber.StatusCreated).JSON(RegisterResponse{
			Message:   "user registered successfully",
			StudentID: req.StudentID,
		})
	}
}

// Login 핸들러: 학번/이름/폰번호가 users에 존재하면 Access/Refresh 발급
// Login godoc
// @Summary      로그인 (학번/이름/전화번호 확인)
// @Description  일치하는 사용자가 있으면 Access/Refresh 토큰 발급, 전화번호란은 숫자만 허용
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        payload body LoginRequest true "로그인 정보"
// @Success      200 {object} LoginResponse
// @Failure      400 {object} ErrorResponse
// @Failure      401 {object} ErrorResponse
// @Router       /auth/login [post]
func Login(d Deps) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// 1) 요청 바디(JSON) 파싱
		var req LoginRequest
		if err := c.BodyParser(&req); err != nil {
			// JSON 파싱 실패 → 400
			return fiber.ErrBadRequest
		}
		// 아주 기초적인 유효성 검사(필드 비었는지 등)
		if req.StudentID == "" || req.Name == "" || req.Phone == "" {
			return fiber.NewError(fiber.StatusBadRequest, "missing required fields: student_id, name, phone_number")
		}
		// 전화번호에 숫자만 들어갔는지 검사
		matched, _ := regexp.MatchString(`^[0-9]+$`, req.Phone)
		if !matched {
			return fiber.NewError(fiber.StatusBadRequest, "only numeric characters are allowed in phone_number")
		}

		// 2) DB에서 존재 여부 확인
		var exists bool
		err := d.DB.QueryRow(c.Context(),
			`SELECT EXISTS(
			   SELECT 1 FROM users
			   WHERE student_id=$1 AND name=$2 AND phone_number=$3
			 )`,
			req.StudentID, req.Name, req.Phone,
		).Scan(&exists)
		if err != nil {
			// DB 에러 → 500
			return fiber.ErrInternalServerError
		}
		if !exists {
			// 사용자 정보가 없으면 → 401
			return fiber.ErrUnauthorized
		}

		// 3) Access 토큰 발급 (짧은 만료, 헤더로 들고 다님)
		access, err := util.IssueAccessToken(req.StudentID)
		if err != nil {
			return fiber.ErrInternalServerError
		}

		// 4) Refresh 토큰 생성 (긴 만료, DB에는 해시만 저장)
		refreshPlain := util.RandomToken(32)        // 안전한 랜덤 바이트 → base64
		hash := sha256.Sum256([]byte(refreshPlain)) // SHA-256 해시(더 강한 KDF도 가능)

		// user agent / ip는 감사성(어디서 발급됐는지 추적)
		ua := string(c.Request().Header.UserAgent())
		ip := clientIP(c)

		// 만료 시각: 환경변수에서 시간(시간 단위)로 읽어와 now() + TTL
		expires := time.Now().Add(time.Hour * time.Duration(util.EnvInt("JWT_REFRESH_TTL_H", 336)))

		// 해시를 base64url로 저장하면 휴먼-리드에도 안전하고 고정 폭에 유리
		hashB64 := base64.RawURLEncoding.EncodeToString(hash[:])

		// DB에 저장(평문 refresh는 절대 저장 X)
		_, err = d.DB.Exec(c.Context(),
			`INSERT INTO auth_refresh_tokens(student_id, token_hash, expires_at, user_agent, ip)
     		 VALUES ($1, $2, $3, $4, $5)
     		 ON CONFLICT (token_hash) DO NOTHING`,
			req.StudentID, hashB64, expires, ua, ip,
		)
		if err != nil {
			return fiber.ErrInternalServerError
		}

		// 5) 클라이언트에 반환
		// 주의: refresh 토큰은 이번 한 번만 평문으로 내려보냄(클라이언트가 보관)
		return c.JSON(LoginResponse{
			AccessToken:  access,
			RefreshToken: refreshPlain,
		})
	}
}

// Refresh 핸들러: refresh 토큰 평문을 받아서 DB의 해시와 비교 후, 새로운 Access 발급
// Refresh godoc
// @Summary      토큰 갱신
// @Description  Refresh 토큰을 사용하여 새로운 Access 토큰을 발급합니다.
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        payload body RefreshRequest true "토큰 갱신 요청 정보"
// @Success      200 {object} RefreshResponse
// @Failure      400 {object} ErrorResponse
// @Failure      401 {object} ErrorResponse
// @Router       /auth/refresh [post]
func Refresh(d Deps) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// 1) 요청 바디(JSON) 파싱
		var req RefreshRequest
		if err := c.BodyParser(&req); err != nil {
			return fiber.NewError(fiber.StatusBadRequest, "invalid request payload")
		}
		if req.RefreshToken == "" {
			return fiber.NewError(fiber.StatusBadRequest, "missing refresh_token")
		}

		// 1.5) Authorization 헤더에서 현재 access token 추출 (블랙리스트용)
		var currentAccessToken string
		if req.AccessToken != "" {
			// 요청 body에서 직접 받은 access_token 사용
			currentAccessToken = req.AccessToken
		} else {
			// 요청 body에 없으면 Authorization 헤더에서 추출
			authHeader := c.Get("Authorization")
			if authHeader != "" && len(authHeader) > 7 && authHeader[:7] == "Bearer " {
				currentAccessToken = authHeader[7:]
			}
		}

		// 2) 평문 refresh → SHA256 → base64url
		hash := sha256.Sum256([]byte(req.RefreshToken))
		hashB64 := base64.RawURLEncoding.EncodeToString(hash[:])

		// 3) DB에서 유효한 리프레시인지 확인 (만료/회수 여부)
		var sid string // student_id
		err := d.DB.QueryRow(c.Context(),
			`SELECT student_id FROM auth_refresh_tokens
			  WHERE token_hash=$1
			    AND revoked_at IS NULL
			    AND now() < expires_at`,
			hashB64,
		).Scan(&sid)
		if err != nil {
			// 토큰이 없거나 만료/회수된 경우
			if err == pgx.ErrNoRows {
				return fiber.ErrUnauthorized // 보안상 구체적인 에러 메시지는 반환하지 않음.
			}
			return fiber.ErrInternalServerError
		}

		// 보안적 측면에서 Refresh 토큰은 1회용으로 설계하는 것이 좋음.
		// 즉, Refresh 시 기존 토큰은 회수(revoke)하고 새 토큰을 발급.
		_, err = d.DB.Exec(c.Context(),
			`UPDATE auth_refresh_tokens
			 SET revoked_at = now()
			 WHERE token_hash = $1`,
			hashB64,
		)
		if err != nil {
			return fiber.ErrInternalServerError
		}

		// 3.5) 이전 access token을 블랙리스트에 추가 (있다면)
		if currentAccessToken != "" {
			if jti, err := util.ExtractJTI(currentAccessToken); err == nil {
				// Redis에 JTI를 블랙리스트로 저장 (TTL은 access token의 만료 시간까지)
				ttlMin := util.EnvInt("JWT_ACCESS_TTL_MIN", 10)
				blacklistKey := "blacklist:" + jti
				_, _ = d.RDB.Set(c.Context(), blacklistKey, "revoked", time.Duration(ttlMin)*time.Minute).Result()
				log.Printf("Added access token to blacklist: %s", jti)
			}
		}

		// 4) 새 Access 발급
		token, err := util.IssueAccessToken(sid)
		if err != nil {
			return fiber.ErrInternalServerError
		}

		// 5) 새 Refresh 토큰 발급 (1회용이므로 새로 발급)
		refreshPlain := util.RandomToken(32)           // 안전한 랜덤 바이트 → base64
		newHash := sha256.Sum256([]byte(refreshPlain)) // SHA-256 해시(더 강한 KDF도 가능)
		newHashB64 := base64.RawURLEncoding.EncodeToString(newHash[:])

		// user agent / ip는 감사성(어디서 발급됐는지 추적)
		ua := string(c.Request().Header.UserAgent())
		ip := clientIP(c)

		// 만료 시각: 환경변수에서 시간(시간 단위)로 읽어와 now() + TTL
		expires := time.Now().Add(time.Hour * time.Duration(util.EnvInt("JWT_REFRESH_TTL_H", 336)))

		// 새 refresh token을 DB에 저장
		_, err = d.DB.Exec(c.Context(),
			`INSERT INTO auth_refresh_tokens(student_id, token_hash, expires_at, user_agent, ip)
     		 VALUES ($1, $2, $3, $4, $5)
     		 ON CONFLICT (token_hash) DO NOTHING`,
			sid, newHashB64, expires, ua, ip,
		)
		if err != nil {
			log.Printf("Refresh: failed to store new refresh token: %v", err)
			return fiber.ErrInternalServerError
		}

		// 6) 클라이언트에 반환
		return c.JSON(RefreshResponse{
			AccessToken:  token,
			RefreshToken: refreshPlain,
		})
	}
}

// clientIP: Fiber 컨텍스트에서 클라이언트 IP를 추출
// - 신뢰 프록시 뒤에 있다면 Fiber Config의 ProxyHeader/TrustedProxy 설정을 고려해야 함
// clientIP godoc
// @summary      클라이언트 IP 추출
// @description  Fiber 컨텍스트에서 클라이언트 IP를 추출합니다.
func clientIP(c *fiber.Ctx) string {
	ip := c.IP()
	if net.ParseIP(ip) == nil {
		log.Printf("Invalid IP address: %s", ip)
		return "0.0.0.0"
	}
	return ip
}

// Logout 핸들러: Access Token과 Refresh Token을 모두 무효화
// Logout godoc
// @Summary      로그아웃
// @Description  현재 사용 중인 Access Token과 Refresh Token을 모두 무효화합니다.
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        Authorization header string false "Bearer {access_token}" default(Bearer )
// @Param        payload body LogoutRequest false "로그아웃 정보 (선택적)"
// @Success      200 {object} LogoutResponse
// @Failure      400 {object} ErrorResponse
// @Failure      401 {object} ErrorResponse
// @Router       /auth/logout [post]
func Logout(d Deps) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// 1) 요청에서 Access Token과 Refresh Token 추출
		var req LogoutRequest
		_ = c.BodyParser(&req) // 에러 무시 (선택적)

		// Access Token 추출 (Authorization 헤더 우선, 없으면 body에서)
		var accessToken string
		authHeader := c.Get("Authorization")
		if authHeader != "" && len(authHeader) > 7 && authHeader[:7] == "Bearer " {
			accessToken = authHeader[7:]
		} else if req.AccessToken != "" {
			accessToken = req.AccessToken
		}

		// 2) Access Token을 블랙리스트에 추가
		if accessToken != "" {
			if jti, err := util.ExtractJTI(accessToken); err == nil {
				// Redis에 JTI를 블랙리스트로 저장
				ttlMin := util.EnvInt("JWT_ACCESS_TTL_MIN", 10)
				blacklistKey := "blacklist:" + jti
				_, err := d.RDB.Set(c.Context(), blacklistKey, "revoked", time.Duration(ttlMin)*time.Minute).Result()
				if err != nil {
					log.Printf("Failed to blacklist access token: %v", err)
				} else {
					log.Printf("Access token blacklisted during logout: %s", jti)
				}
			}
		}

		// 3) Refresh Token 무효화 (DB에서 revoke)
		if req.RefreshToken != "" {
			// 평문 refresh → SHA256 → base64url
			hash := sha256.Sum256([]byte(req.RefreshToken))
			hashB64 := base64.RawURLEncoding.EncodeToString(hash[:])

			// DB에서 해당 refresh token을 revoke
			result, err := d.DB.Exec(c.Context(),
				`UPDATE auth_refresh_tokens
				 SET revoked_at = now()
				 WHERE token_hash = $1 AND revoked_at IS NULL`,
				hashB64,
			)
			if err != nil {
				log.Printf("Failed to revoke refresh token: %v", err)
			} else {
				rowsAffected := result.RowsAffected()
				if rowsAffected > 0 {
					log.Printf("Refresh token revoked during logout")
				}
			}
		}

		// 4) 성공 응답
		return c.JSON(LogoutResponse{
			Message: "logged out successfully",
		})
	}
}

// LogoutAll 핸들러: 해당 사용자의 모든 세션(모든 디바이스)을 무효화
// LogoutAll godoc
// @Summary      전체 로그아웃 (모든 디바이스)
// @Description  현재 사용자의 모든 Refresh Token을 무효화하여 모든 디바이스에서 로그아웃합니다.
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        Authorization header string true "Bearer {access_token}" default(Bearer )
// @Success      200 {object} LogoutResponse
// @Failure      401 {object} ErrorResponse
// @Router       /auth/logout-all [post]
func LogoutAll(d Deps) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// JWT 미들웨어에서 저장한 학번(sub) - 이 핸들러는 인증이 필요함
		studentID, ok := c.Locals("student_id").(string)
		if !ok || studentID == "" {
			return fiber.ErrUnauthorized
		}

		// 1) 현재 Access Token을 블랙리스트에 추가
		authHeader := c.Get("Authorization")
		if authHeader != "" && len(authHeader) > 7 && authHeader[:7] == "Bearer " {
			accessToken := authHeader[7:]
			if jti, err := util.ExtractJTI(accessToken); err == nil {
				ttlMin := util.EnvInt("JWT_ACCESS_TTL_MIN", 10)
				blacklistKey := "blacklist:" + jti
				_, err := d.RDB.Set(c.Context(), blacklistKey, "revoked", time.Duration(ttlMin)*time.Minute).Result()
				if err != nil {
					log.Printf("Failed to blacklist current access token: %v", err)
				}
			}
		}

		// 2) 해당 사용자의 모든 Refresh Token을 revoke
		result, err := d.DB.Exec(c.Context(),
			`UPDATE auth_refresh_tokens
			 SET revoked_at = now()
			 WHERE student_id = $1 AND revoked_at IS NULL`,
			studentID,
		)
		if err != nil {
			log.Printf("Failed to revoke all refresh tokens for user %s: %v", studentID, err)
			return fiber.ErrInternalServerError
		}

		rowsAffected := result.RowsAffected()
		log.Printf("Revoked %d refresh tokens for user %s (logout-all)", rowsAffected, studentID)

		// 3) 성공 응답
		return c.JSON(LogoutResponse{
			Message: "logged out from all devices successfully",
		})
	}
}

/*
[보안 팁]
- Refresh 토큰은 탈취 시 심각한 위험 → https 쿠키(httponly/secure) 보관을 고려.
- 로그에 토큰 평문을 남기지 않기. 에러 메시지도 토큰/전화번호 등 민감정보 포함 금지.
- 다중 디바이스 로그아웃/전체 로그아웃: auth_refresh_tokens에서 student_id로 revoked_at 업데이트.
*/

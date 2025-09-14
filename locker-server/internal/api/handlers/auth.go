package handlers

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/binary"
	"fmt" // 추가
	"log"
	"net"
	"regexp"
	"strings" // 추가
	"time"

	"github.com/KUCSEPotato/locker-server/internal/util"
	"github.com/gofiber/fiber/v2"

	// "github.com/golang-jwt/jwt/v5"
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

type GetMeResponse struct {
	StudentID string `json:"student_id"`
	Name      string `json:"name"`
	Phone     string `json:"phone_number"`
}

// LoginOrRegisterResponse is the response returned by LoginOrRegister handler
type LoginOrRegisterRequest struct {
	StudentID string `json:"student_id"`
	Name      string `json:"name"`
	Phone     string `json:"phone_number"`
}

// LoginOrRegisterResponse is the response returned by LoginOrRegister handler
type LoginOrRegisterResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	SerialID     int    `json:"serial_id"`
}

/*
-- users 테이블에 custom_serial 컬럼이 없다면 추가
ALTER TABLE users
  ADD COLUMN IF NOT EXISTS custom_serial BIGINT;

-- (권장) 동일인 판정 유니크 키: 완전 일치(학번, 이름, 전화번호)
CREATE UNIQUE INDEX IF NOT EXISTS ux_users_ident
  ON users (student_id, name, phone_number);

-- (선택) 조회 최적화
CREATE INDEX IF NOT EXISTS ix_users_custom_serial ON users(custom_serial);
*/

// LoginOrRegister godoc
// @Summary      로그인 또는 자동 회원가입 (통합 인증)
// @Description  학번/이름/전화번호가 일치하면 로그인, 불일치하면 새로 회원가입 후 로그인.
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        payload body LoginOrRegisterRequest true "로그인/회원가입 정보"
// @Success      200 {object} LoginOrRegisterResponse "기존 사용자 로그인 성공"
// @Success      201 {object} LoginOrRegisterResponse "새 사용자 회원가입 및 로그인 성공"
// @Failure      400 {object} ErrorResponse "missing required fields: student_id, name, phone_number"
// @Failure      400 {object} ErrorResponse "invalid student_id format"
// @Failure      400 {object} ErrorResponse "invalid phone_number format"
// @Failure      400 {object} ErrorResponse "only numeric characters are allowed in phone_number"
// @Failure      400 {object} ErrorResponse "invalid name length"
// @Failure      500 {object} ErrorResponse "internal server error"
// @Router       /auth/login-or-register [post]
func LoginOrRegister(d Deps) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// 1) 요청 파싱
		var req LoginOrRegisterRequest
		if err := c.BodyParser(&req); err != nil {
			return fiber.ErrBadRequest
		}
		req.StudentID = strings.TrimSpace(req.StudentID)
		req.Name = strings.TrimSpace(req.Name)
		req.Phone = strings.TrimSpace(req.Phone)

		// 2) 기본 검증
		if req.StudentID == "" || req.Name == "" || req.Phone == "" {
			return fiber.NewError(fiber.StatusBadRequest, "missing required fields: student_id, name, phone_number")
		}
		if ok := regexp.MustCompile(`^\d{10}$`).MatchString(req.StudentID); !ok {
			return fiber.NewError(fiber.StatusBadRequest, "invalid student_id format")
		}
		if len(req.Phone) < 10 || len(req.Phone) > 15 {
			return fiber.NewError(fiber.StatusBadRequest, "invalid phone_number format")
		}
		if ok := regexp.MustCompile(`^\d+$`).MatchString(req.Phone); !ok {
			return fiber.NewError(fiber.StatusBadRequest, "only numeric characters are allowed in phone_number")
		}
		if l := len([]rune(req.Name)); l < 2 || l > 20 {
			return fiber.NewError(fiber.StatusBadRequest, "invalid name length")
		}

		// 3) 커스텀 일련번호 생성 (학번+전화번호+salt → SHA256 → 12자리 숫자)
		customSerial, err := generateCustomSerial(req.StudentID, req.Name, req.Phone)
		if err != nil {
			log.Printf("GenerateCustomSerial failed: %v", err)
			return fiber.ErrInternalServerError
		}

		// 4) 원자적 UPSERT: (student_id, name, phone_number) 유니크 기준
		//    - 새 레코드면 201, 기존이면 200
		var (
			serialID int64
			inserted bool
		)
		err = d.DB.QueryRow(c.Context(), `
			INSERT INTO users (student_id, name, phone_number, serial_id, created_at)
			VALUES ($1, $2, $3, $4, now())
			ON CONFLICT (student_id, name, phone_number)
			DO UPDATE SET
    		name = EXCLUDED.name,
    		phone_number = EXCLUDED.phone_number,
			updated_at = now()
			RETURNING serial_id, (xmax = 0) AS inserted
		`, req.StudentID, req.Name, req.Phone, customSerial).Scan(&serialID, &inserted)
		if err != nil {
			log.Printf("LoginOrRegister: upsert users failed: %v", err)
			return fiber.ErrInternalServerError
		}

		statusCode := fiber.StatusOK
		if inserted {
			statusCode = fiber.StatusCreated
			log.Printf("New user registered: student_id=%s, name=%s", req.StudentID, req.Name)
		} else {
			log.Printf("Existing user logged in: student_id=%s", req.StudentID)
		}

		// 5) Access/Refresh 토큰 발급
		accessToken, err := util.IssueAccessToken(serialID, req.StudentID)
		if err != nil {
			log.Printf("LoginOrRegister: failed to issue access token for student_id=%s: %v", req.StudentID, err)
			return fiber.ErrInternalServerError
		}

		refreshPlain := util.RandomToken(32)
		refreshHash := sha256.Sum256([]byte(refreshPlain))
		hashB64 := base64.RawURLEncoding.EncodeToString(refreshHash[:])

		userAgent := string(c.Request().Header.UserAgent())
		clientIPAddr := clientIP(c)
		refreshExpires := time.Now().Add(time.Hour * time.Duration(util.EnvInt("JWT_REFRESH_TTL_H", 336)))

		_, err = d.DB.Exec(c.Context(), `
			INSERT INTO auth_refresh_tokens (user_serial_id, token_hash, expires_at, user_agent, ip)
			VALUES ($1, $2, $3, $4, $5)
			ON CONFLICT (token_hash) DO NOTHING
		`, serialID, hashB64, refreshExpires, userAgent, clientIPAddr)
		if err != nil {
			log.Printf("LoginOrRegister: failed to store refresh token for user with serial_id=%d: %v", serialID, err)
			return fiber.ErrInternalServerError
		}

		// 6) 응답
		return c.Status(statusCode).JSON(LoginOrRegisterResponse{
			AccessToken:  accessToken,
			RefreshToken: refreshPlain, // 평문은 이 한 번만 반환
			SerialID:     int(serialID),
		})
	}
}

// ───────────────────────────────────────────────────────────────────────────────
// Helpers
// ───────────────────────────────────────────────────────────────────────────────

// generateCustomSerial
// 학번 + 전화번호 + "ku_info" 를 입력으로 SHA256 해시 → 상위 8바이트를 숫자로 변환 → 12자리로 축소(모듈러)
func generateCustomSerial(studentID, name, phone string) (int64, error) {
	id := strings.TrimSpace(studentID)
	nm := strings.TrimSpace(name)
	pn := strings.TrimSpace(phone)
	if id == "" || nm == "" || pn == "" {
		return 0, fmt.Errorf("invalid studentID or name or phone")
	}
	const salt = "ku_info"
	input := id + nm + pn + salt
	sum := sha256.Sum256([]byte(input))

	// 앞 8바이트 → uint64
	u := binary.BigEndian.Uint64(sum[:8])

	// 12자리 숫자로 축소 (0 ~ 999,999,999,999)
	const base uint64 = 1_000_000_000_000
	num := u % base
	return int64(num), nil
}

/*
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        payload body RegisterRequest true "회원가입 정보"
// @Success      201 {object} RegisterResponse
// @Failure      400 {object} ErrorResponse "missing required fields: student_id, name, phone_number"
// @Failure      400 {object} ErrorResponse "only numeric characters are allowed in phone_number"
// @Failure      409 {object} ErrorResponse "student_id already exists"
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
// @Failure      400 {object} ErrorResponse "missing required fields: student_id, name, phone_number"
// @Failure      400 {object} ErrorResponse "only numeric characters are allowed in phone_number"
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
*/

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
		var sid int64 // user_serial_id
		err := d.DB.QueryRow(c.Context(),
			`SELECT user_serial_id FROM auth_refresh_tokens
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

		// 3.2) serial_id로 student_id 조회
		var studentID string
		err = d.DB.QueryRow(c.Context(), `SELECT student_id FROM users WHERE serial_id = $1`, sid).Scan(&studentID)
		if err != nil {
			log.Printf("Refresh: could not find user with serial_id %d: %v", sid, err)
			return fiber.ErrUnauthorized
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
		token, err := util.IssueAccessToken(sid, studentID)
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
			`INSERT INTO auth_refresh_tokens(user_serial_id, token_hash, expires_at, user_agent, ip)
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

// GetMe 핸들러: 현재 로그인된 사용자의 정보 조회
// GetMe godoc
// @Summary      현재 로그인된 사용자 정보 조회
// @Description  JWT 토큰을 통해 인증된 현재 사용자의 학번, 이름, 전화번호를 반환합니다.
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        Authorization header string true "Bearer {access_token}" default(Bearer )
// @Success      200 {object} GetMeResponse
// @Failure      401 {object} ErrorResponse "unauthorized - invalid or missing token"
// @Failure      404 {object} ErrorResponse "user not found"
// @Failure      500 {object} ErrorResponse "internal server error"
// @Router       /auth/me [get]
func GetMe(d Deps) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// JWT 미들웨어에서 저장한 학번(sub) - 이 핸들러는 인증이 필요함
		serialID, ok := c.Locals("user_serial_id").(int64)
		if !ok || serialID == 0 {
			return fiber.ErrUnauthorized
		}

		// DB에서 해당 사용자 정보 조회
		var studentID, name, phone string
		err := d.DB.QueryRow(c.Context(),
			`SELECT student_id, name, phone_number FROM users WHERE serial_id = $1 LIMIT 1`,
			serialID,
		).Scan(&studentID, &name, &phone)

		if err != nil {
			if err == pgx.ErrNoRows {
				return fiber.NewError(fiber.StatusNotFound, "user not found")
			}
			log.Printf("GetMe: failed to query user info: %v", err)
			return fiber.ErrInternalServerError
		}

		// 사용자 정보 반환
		return c.JSON(GetMeResponse{
			StudentID: studentID,
			Name:      name,
			Phone:     phone,
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
// @Param        payload body LogoutRequest false "로그아웃 정보"
// @Success      200 {object} LogoutResponse
// @Failure      400 {object} ErrorResponse
// @Failure      401 {object} ErrorResponse
// @Router       /auth/logout [post]
func Logout(d Deps) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var req LogoutRequest
		_ = c.BodyParser(&req)

		req.AccessToken = strings.TrimSpace(req.AccessToken)
		req.RefreshToken = strings.TrimSpace(req.RefreshToken)

		// 미들웨어로부터 인증된 사용자 확인
		authenticatedSerialID, _ := c.Locals("user_serial_id").(int64)

		// 인증된 사용자도 없고 refresh token도 없으면 거부
		if authenticatedSerialID == 0 && req.RefreshToken == "" {
			return fiber.ErrUnauthorized
		}

		// 1) access token 블랙리스트 처리 (가능하면)
		accessToken := ""
		authHeader := c.Get("Authorization")
		if strings.HasPrefix(authHeader, "Bearer ") {
			accessToken = strings.TrimPrefix(authHeader, "Bearer ")
		} else if req.AccessToken != "" {
			accessToken = req.AccessToken
		}
		if accessToken != "" {
			// 우선 JTI 추출 시도
			if jti, err := util.ExtractJTI(accessToken); err == nil && jti != "" {
				ttlMin := util.EnvInt("JWT_ACCESS_TTL_MIN", 10)
				_, _ = d.RDB.Set(c.Context(), "blacklist:"+jti, "revoked", time.Duration(ttlMin)*time.Minute).Result()
			} else {
				// JTI가 없거나 추출 실패 시 토큰 해시로 블랙리스트
				h := sha256.Sum256([]byte(accessToken))
				key := "blacklist:token:" + base64.RawURLEncoding.EncodeToString(h[:])
				ttlMin := util.EnvInt("JWT_ACCESS_TTL_MIN", 10)
				_, _ = d.RDB.Set(c.Context(), key, "revoked", time.Duration(ttlMin)*time.Minute).Result()
			}
		}

		// 2) refresh token이 제공된 경우 해당 토큰만 revoke
		if req.RefreshToken != "" {
			hash := sha256.Sum256([]byte(req.RefreshToken))
			hashB64 := base64.RawURLEncoding.EncodeToString(hash[:])

			tag, err := d.DB.Exec(c.Context(),
				`UPDATE auth_refresh_tokens
                 SET revoked_at = now()
                 WHERE token_hash = $1 AND revoked_at IS NULL`,
				hashB64,
			)
			if err != nil {
				log.Printf("Logout: failed to revoke refresh token by hash: %v", err)
				return fiber.ErrInternalServerError
			}
			if tag.RowsAffected() > 0 {
				_, _ = d.RDB.Del(c.Context(), "refresh_token:"+req.RefreshToken).Result()
			}
			return c.JSON(LogoutResponse{Message: "logged out successfully"})
		}

		// 3) refresh token 미제공이면서 인증된 사용자가 있으면 해당 사용자의 모든 refresh 토큰 revoke
		if authenticatedSerialID != 0 {
			tag, err := d.DB.Exec(c.Context(),
				`UPDATE auth_refresh_tokens
                 SET revoked_at = now()
                 WHERE user_serial_id = $1 AND revoked_at IS NULL`,
				authenticatedSerialID,
			)
			if err != nil {
				log.Printf("Logout: failed to revoke refresh tokens for user %d: %v", authenticatedSerialID, err)
				return fiber.ErrInternalServerError
			}
			log.Printf("Logout: revoked %d refresh tokens for user %d", tag.RowsAffected(), authenticatedSerialID)
			return c.JSON(LogoutResponse{Message: "logged out successfully"})
		}

		return fiber.ErrUnauthorized
	}
}

/*
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
*/
/*
[보안 팁]
- Refresh 토큰은 탈취 시 심각한 위험 → https 쿠키(httponly/secure) 보관을 고려.
- 로그에 토큰 평문을 남기지 않기. 에러 메시지도 토큰/전화번호 등 민감정보 포함 금지.
- 다중 디바이스 로그아웃/전체 로그아웃: auth_refresh_tokens에서 student_id로 revoked_at 업데이트.
*/

// Login
// @Summary      로그인 (학번/이름/전화번호 확인)
// @Description  일치하는 사용자가 있으면 Access/Refresh 토큰 발급
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        payload body struct{StudentID string `json:"student_id"`; Name string `json:"name"`; Phone string `json:"phone_number"`} true "로그인 정보"
// @Success      200 {object} map[string]string "access_token, refresh_token"
// @Failure      400 {object} map[string]any
// @Failure      401 {object} map[string]any
// @Router       /auth/login [post]
package handlers

import (
	"crypto/sha256"
	"encoding/base64"
	"net"
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
	RefreshToken string `json:"refresh_token"`
}

// 취약점 주의: hardcoded-credentials Embedding credentials in source code risks unauthorized access
type RefreshResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

// Login 핸들러: 학번/이름/폰번호가 users에 존재하면 Access/Refresh 발급
// Login godoc
// @Summary      로그인 (학번/이름/전화번호 확인)
// @Description  일치하는 사용자가 있으면 Access/Refresh 토큰 발급
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
		if req.Name == "" || req.Phone == "" {
			return fiber.NewError(fiber.StatusBadRequest, "missing fields")
		}

		// 아주 기초적인 유효성 검사(필드 비었는지 등)
		if req.StudentID == "" || req.Name == "" || req.Phone == "" {
			return fiber.NewError(fiber.StatusBadRequest, "missing fields")
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
			 VALUES ($1,$2,$3,$4,$5)`,
			req.StudentID, hashB64, expires, ua, ip,
		)
		if err != nil {
			return fiber.ErrInternalServerError
		}

		// 5) 클라이언트에 반환
		// 주의: refresh 토큰은 이번 한 번만 평문으로 내려보냄(클라이언트가 보관)
		return c.JSON(fiber.Map{
			"access_token":  access,
			"refresh_token": refreshPlain,
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
		var req struct {
			Refresh string `json:"refresh_token"`
		}
		if err := c.BodyParser(&req); err != nil {
			return fiber.ErrBadRequest
		}
		if req.Refresh == "" {
			return fiber.NewError(fiber.StatusBadRequest, "missing refresh_token")
		}

		// 2) 평문 refresh → SHA256 → base64url
		hash := sha256.Sum256([]byte(req.Refresh))
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
				return fiber.ErrUnauthorized
			}
			return fiber.ErrInternalServerError
		}

		// 4) 새 Access 발급
		token, err := util.IssueAccessToken(sid)
		if err != nil {
			return fiber.ErrInternalServerError
		}
		return c.JSON(fiber.Map{"access_token": token})
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
		return ""
	}
	return ip
}

/*
[보안 팁]
- Refresh 토큰은 탈취 시 심각한 위험 → https 쿠키(httponly/secure) 보관을 고려.
- 로그에 토큰 평문을 남기지 않기. 에러 메시지도 토큰/전화번호 등 민감정보 포함 금지.
- 다중 디바이스 로그아웃/전체 로그아웃: auth_refresh_tokens에서 student_id로 revoked_at 업데이트.
*/

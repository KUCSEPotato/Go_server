package middleware

import (
	"log"
	"os"
	"strings"

	"github.com/KUCSEPotato/locker-server/internal/util"
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

// Deps: 미들웨어에서 사용할 의존성
type Deps struct {
	DB  *pgxpool.Pool
	RDB *redis.Client
}

// JWTAuth 는 보호된 라우트에서 사용되는 미들웨어로,
// 1) Authorization 헤더에 Bearer 토큰이 있는지 확인
// 2) 토큰 서명/클레임(iss, aud, exp 등) 검증
// 3) 블랙리스트 체크
// 4) sub(학번)를 c.Locals("student_id")에 저장해 핸들러에서 사용 가능하게 함
func JWTAuth(d Deps) fiber.Handler {
	// 환경변수로부터 검증에 필요한 값 로드
	secret := []byte(os.Getenv("JWT_ACCESS_SECRET"))
	iss := os.Getenv("JWT_ISS")
	aud := os.Getenv("JWT_AUD")

	// 환경변수 검증
	if secret == nil || iss == "" || aud == "" {
		log.Fatal("JWT environment variables are not set properly")
	}

	return func(c *fiber.Ctx) error {
		// HTTP Authorization 헤더에서 Bearer 토큰 추출
		// 예) "Authorization: Bearer eyJhbGciOi..."
		authz := c.Get("Authorization")
		if !strings.HasPrefix(authz, "Bearer ") {
			// 토큰이 없거나 포맷이 잘못되면 401
			return fiber.ErrUnauthorized
		}
		tokenStr := strings.TrimPrefix(authz, "Bearer ")

		// 블랙리스트 체크 (먼저 체크해서 불필요한 파싱 방지)
		if jti, err := util.ExtractJTI(tokenStr); err == nil {
			blacklistKey := "blacklist:" + jti
			if exists, _ := d.RDB.Exists(c.Context(), blacklistKey).Result(); exists > 0 {
				log.Printf("Token is blacklisted: %s", jti)
				return fiber.ErrUnauthorized
			}
		}

		// jwt.Parse: 토큰 구조/서명/표준 클레임을 검증.
		// - keyfunc: 어떤 키로 서명했는지를 서버가 결정 (여기서는 HS256 + secret)
		// - WithIssuer/WithAudience: iss/aud 값을 체크
		token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (any, error) {
			// alg 고정 방어: 서버가 허용한 알고리즘(HS256)만 받는다.
			if t.Method.Alg() != jwt.SigningMethodHS256.Alg() {
				return nil, fiber.ErrUnauthorized
			}
			return secret, nil
		}, jwt.WithIssuer(iss), jwt.WithAudience(aud))
		if err != nil {
			log.Print("Failed to parse JWT token")
			return fiber.ErrUnauthorized
		}
		if !token.Valid {
			log.Print("Invalid JWT token")
			return fiber.ErrUnauthorized
		}

		// payload(클레임) 꺼내기
		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			return fiber.ErrUnauthorized
		}

		// sub(주체) = 학번. 핸들러에서 c.Locals("student_id")로 꺼내씀.
		sub, _ := claims["sub"].(string)
		if sub == "" {
			log.Print("Missing student_id claim")
			return fiber.ErrUnauthorized
		}
		c.Locals("student_id", sub)

		// 다음 미들웨어/핸들러 실행
		return c.Next()
	}
}

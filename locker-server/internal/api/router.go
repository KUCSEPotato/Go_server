package api

import (
	"github.com/gofiber/fiber/v2"

	// 실제 요청을 처리하는 핸들러
	"github.com/KUCSEPotato/locker-server/internal/api/handlers"
	// JWT 인증 미들웨어
	"github.com/KUCSEPotato/locker-server/internal/api/middleware"
	// fiberSwagger "github.com/swaggo/fiber-swagger"
)

// Setup 함수는 main에서 호출되어 라우터 트리를 구성한다.
func Setup(app *fiber.App, deps handlers.Deps) {
	// 최상위 prefix: /api
	api := app.Group("/api")
	// 버전 그룹: /api/v1
	v1 := api.Group("/v1")

	// --- 인증(로그인/리프레시) 엔드포인트는 공개(public) ---
	v1.Post("/auth/login", handlers.Login(deps))     // 학번/이름/폰번호 확인 → 토큰 발급
	v1.Post("/auth/refresh", handlers.Refresh(deps)) // 리프레시 토큰으로 액세스 갱신

	// [250904] 추가: 헬스 체크 엔드포인트
	// --- 헬스 체크 엔드포인트 ---
	v1.Get("/health", handlers.HealthCheck(deps.DB, deps.RDB)) // DB, Redis 상태 확인

	// --- 아래부터는 JWT가 있어야 접근 가능한 보호 API ---
	// 빈 prefix("")에 JWT 미들웨어를 덧씌워서 같은 그룹 안 라우트에 공통적용
	authed := v1.Group("", middleware.JWTAuth())

	authed.Get("/lockers", handlers.ListLockers(deps))                // 사물함 목록 조회
	authed.Get("/lockers/me", handlers.GetMyLocker(deps))             // <-- 추가
	authed.Post("/lockers/:id/hold", handlers.HoldLocker(deps))       // 사물함 홀드(선점)
	authed.Post("/lockers/:id/confirm", handlers.ConfirmLocker(deps)) // 확정
	authed.Post("/lockers/:id/release", handlers.ReleaseLocker(deps)) // 해제

	// swagger
	// app.Get("/swagger/*", fiberSwagger.WrapHandler)
}

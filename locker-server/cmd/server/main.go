// @title           Locker Reservation API
// @version         1.0
// @description     사물함 선착순 예약 시스템의 백엔드 API 문서
// @BasePath        /api/v1

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Bearer {access_token}
package main

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/gofiber/fiber/v2"
	// logger: HTTP 요청/응답을 콘솔에 기록하는 미들웨어 (개발 시 매우 유용)
	"github.com/gofiber/fiber/v2/middleware/logger"
	// recover: 핸들러 내부에서 panic이 나도 서버가 죽지 않게 막아주는 미들웨어
	"github.com/gofiber/fiber/v2/middleware/recover"

	"github.com/KUCSEPotato/locker-server/internal/api"
	"github.com/KUCSEPotato/locker-server/internal/api/handlers"
	"github.com/KUCSEPotato/locker-server/internal/cache"
	"github.com/KUCSEPotato/locker-server/internal/db"

	// .env 자동 로딩
	"github.com/joho/godotenv"

	// swagger
	_ "github.com/KUCSEPotato/locker-server/docs"
	fiberSwagger "github.com/swaggo/fiber-swagger" // fiberSwagger
)

func main() {
	// env 파일 로딩
	// 상대 경로: ../configs/.env
	_ = godotenv.Load("/Users/potato/Desktop/Dev/Go_server/locker-server/configs/.env")

	// 서버의 표준 시간대(로그 타임스탬프 등)를 서울로 고정
	_ = os.Setenv("TZ", "Asia/Seoul")

	// 컨텍스트: DB 커넥션 준비/헬스체크 등에 사용
	ctx := context.Background()

	// PostgreSQL 풀 생성 (pgxpool)
	pool := db.NewPool(ctx)

	// Redis 클라이언트 생성 (원자적 hold, 레이트리밋 등에 사용)
	rdb := cache.NewRedis()

	// Fiber 앱 생성 + 핵심 타임아웃 설정
	app := fiber.New(fiber.Config{
		AppName:      os.Getenv("APP_NAME"),
		ReadTimeout:  5 * time.Second,  // 요청 바디 읽기 제한
		WriteTimeout: 5 * time.Second,  // 응답 쓰기 제한
		IdleTimeout:  30 * time.Second, // Keep-Alive 대기 시간
		// BodyLimit, ProxyHeader 등도 상황에 따라 추가 가능
	})

	// 전역 미들웨어 장착
	app.Use(
		logger.New(),  // 요청 로그 출력
		recover.New(), // panic 복구
	)

	// 의존성 주입용 구조체(핸들러들이 DB/Redis에 접근할 때 사용)
	deps := handlers.Deps{DB: pool, RDB: rdb}

	// 라우팅 트리 구성
	api.Setup(app, deps)

	// swagger
	app.Get("/swagger/*", fiberSwagger.WrapHandler)

	// HTTP 서버 시작 (예: :3000)
	log.Fatal(app.Listen(os.Getenv("APP_ADDR")))
}

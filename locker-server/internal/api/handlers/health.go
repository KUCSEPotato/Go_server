// [250904 추가]
package handlers

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

// HealthCheck 핸들러: 간단한 헬스 체크
// HealthCheck godoc
// @Summary      헬스 체크
// @Description  서버 상태 확인 (DB, Redis 연결 상태 포함)
// @Tags         health
// @Accept       json
// @Produce      json
// @Success      200 {object} map[string]string "서버가 정상입니다"
// @Failure      500 {object} map[string]string "서버에 문제가 있습니다"
// @Router       /health [get]
func HealthCheck(db *pgxpool.Pool, rdb *redis.Client) fiber.Handler {
	return func(c *fiber.Ctx) error {
		ctx, cancel := context.WithTimeout(c.Context(), 2*time.Second)
		defer cancel()

		// PostgreSQL 연결 상태 확인
		if err := db.Ping(ctx); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"status": "unhealthy",
				"db":     "unreachable",
				"error":  err.Error(),
			})
		}

		// Redis 연결 상태 확인
		if err := rdb.Ping(ctx).Err(); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"status": "unhealthy",
				"redis":  "unreachable",
				"error":  err.Error(),
			})
		}

		// 모든 체크 통과 시
		return c.JSON(fiber.Map{
			"status": "healthy",
			"db":     "reachable",
			"redis":  "reachable",
		})
	}
}

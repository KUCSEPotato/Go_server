package scheduler

import (
	"log"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

// CleanupExpiredHolds 만료된 hold를 주기적으로 expired로 변경
func StartCleanupScheduler(db *pgxpool.Pool, rdb *redis.Client) {
	ticker := time.NewTicker(10 * time.Second) // 10초마다 실행 (fallback용)
	go func() {
		defer ticker.Stop()
		for range ticker.C {
			CheckAndCleanupAllExpiredHolds(db, rdb)
		}
	}()
	log.Println("Cleanup scheduler started: checking expired holds every 10 seconds (fallback)")
}

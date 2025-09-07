package scheduler

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

// StartRealtimeCleanup Redis keyspace notifications를 사용한 실시간 cleanup
func StartRealtimeCleanup(db *pgxpool.Pool, rdb *redis.Client) {
	// Redis keyspace notifications 활성화
	_, err := rdb.ConfigSet(context.Background(), "notify-keyspace-events", "Ex").Result()
	if err != nil {
		log.Printf("Failed to enable Redis keyspace notifications: %v", err)
		return
	}

	// expire 이벤트 구독
	pubsub := rdb.PSubscribe(context.Background(), "__keyevent@0__:expired")
	defer pubsub.Close()

	log.Println("Real-time cleanup started: listening for Redis key expiration events")

	go func() {
		for msg := range pubsub.Channel() {
			// 메시지 형태: "locker:hold:101"
			if strings.HasPrefix(msg.Payload, "locker:hold:") {
				// locker ID 추출
				parts := strings.Split(msg.Payload, ":")
				if len(parts) == 3 {
					lockerIDStr := parts[2]
					lockerID, err := strconv.Atoi(lockerIDStr)
					if err != nil {
						log.Printf("Invalid locker ID in expired key: %s", msg.Payload)
						continue
					}

					// DB에서 해당 locker의 hold 상태를 expired로 업데이트
					if err := markHoldAsExpired(db, lockerID); err != nil {
						log.Printf("Failed to mark locker %d as expired: %v", lockerID, err)
					} else {
						log.Printf("Successfully marked locker %d hold as expired (real-time)", lockerID)
					}
				}
			}
		}
	}()
}

// markHoldAsExpired 특정 locker의 hold 상태를 expired로 변경
func markHoldAsExpired(db *pgxpool.Pool, lockerID int) error {
	query := `
		UPDATE locker_assignments 
		SET state = 'expired' 
		WHERE locker_id = $1 AND state = 'hold'`

	result, err := db.Exec(context.Background(), query, lockerID)
	if err != nil {
		return fmt.Errorf("failed to update locker %d: %w", lockerID, err)
	}

	rowsAffected := result.RowsAffected()
	if rowsAffected == 0 {
		log.Printf("No hold record found for locker %d (may have been already processed)", lockerID)
	}

	return nil
}

package scheduler

import (
	"context"
	"log"
	"strconv"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

// CheckAndCleanupExpiredHold API 요청 시 특정 locker의 만료된 hold를 체크하고 정리
func CheckAndCleanupExpiredHold(db *pgxpool.Pool, rdb *redis.Client, lockerID int) error {
	ctx := context.Background()
	redisKey := "locker:hold:" + strconv.Itoa(lockerID)

	// Redis에서 키가 존재하는지 확인
	exists, err := rdb.Exists(ctx, redisKey).Result()
	if err != nil {
		log.Printf("Failed to check Redis key existence for locker %d: %v", lockerID, err)
		return err
	}

	// Redis에 키가 없으면 DB의 hold 상태를 expired로 변경
	if exists == 0 {
		query := `
			UPDATE locker_assignments 
			SET state = 'expired' 
			WHERE locker_id = $1 AND state = 'hold'`

		result, err := db.Exec(ctx, query, lockerID)
		if err != nil {
			log.Printf("Failed to mark expired hold for locker %d: %v", lockerID, err)
			return err
		}

		rowsAffected := result.RowsAffected()
		if rowsAffected > 0 {
			log.Printf("Marked expired hold for locker %d during API call", lockerID)
		}
	}

	return nil
}

// CheckAndCleanupAllExpiredHolds 모든 만료된 hold를 체크하고 정리 (API 요청과 무관하게)
func CheckAndCleanupAllExpiredHolds(db *pgxpool.Pool, rdb *redis.Client) error {
	ctx := context.Background()

	// DB에서 현재 hold 상태인 모든 locker 조회
	rows, err := db.Query(ctx, `
		SELECT locker_id 
		FROM locker_assignments 
		WHERE state = 'hold'`)
	if err != nil {
		return err
	}
	defer rows.Close()

	var cleanedCount int
	for rows.Next() {
		var lockerID int
		if err := rows.Scan(&lockerID); err != nil {
			continue
		}

		// Redis에서 해당 키가 존재하는지 확인
		redisKey := "locker:hold:" + strconv.Itoa(lockerID)
		exists, err := rdb.Exists(ctx, redisKey).Result()
		if err != nil {
			log.Printf("Failed to check Redis key for locker %d: %v", lockerID, err)
			continue
		}

		// Redis에 키가 없으면 만료된 것으로 간주하고 DB 업데이트
		if exists == 0 {
			_, err := db.Exec(ctx, `
				UPDATE locker_assignments 
				SET state = 'expired' 
				WHERE locker_id = $1 AND state = 'hold'`, lockerID)
			if err != nil {
				log.Printf("Failed to mark expired hold for locker %d: %v", lockerID, err)
				continue
			}
			cleanedCount++
		}
	}

	if cleanedCount > 0 {
		log.Printf("Cleaned up %d expired holds", cleanedCount)
	}

	return nil
}

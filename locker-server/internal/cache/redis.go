package cache

import (
	"os"

	"github.com/redis/go-redis/v9"
)

// NewRedis: REDIS_ADDR 환경변수(예: "localhost:6379")로 클라이언트 생성
func NewRedis() *redis.Client {
	addr := os.Getenv("REDIS_ADDR")
	if addr == "" {
		addr = "localhost:6379"
	}
	return redis.NewClient(&redis.Options{
		Addr: addr,
		// Password: "", DB: 0 등도 상황에 따라 지정
	})
}

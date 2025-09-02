package db

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// NewPool: 환경변수 DB_URL을 읽어 커넥션 풀을 생성하고 Ping으로 연결 확인
// - 실제 운영에서는 MaxConns, MinConns, 연결 타임아웃 등을 환경변수로 빼서 조정하는 것을 권장
func NewPool(ctx context.Context) *pgxpool.Pool {
	url := os.Getenv("DB_URL")
	if url == "" {
		log.Fatal("DB_URL is empty")
	}

	cfg, err := pgxpool.ParseConfig(url)
	if err != nil {
		log.Fatalf("pgx ParseConfig: %v", err)
	}

	// 풀 사이즈(예시). 워크로드/DB 서버 사양/쿼리 특성에 맞춰 조절 필요.
	cfg.MaxConns = 10

	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		log.Fatalf("pgxpool NewWithConfig: %v", err)
	}

	// Ping: 실제 연결 가능 여부 확인 (3초 제한)
	ctxPing, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	if err := pool.Ping(ctxPing); err != nil {
		log.Fatalf("pgxpool Ping: %v", err)
	}
	return pool
}

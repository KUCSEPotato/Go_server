run:
	APP_ADDR=:3000 go run ./cmd/server

migrate:
	# Postgres 컨테이너에 연결해서 초기 스키마(0001_init.sql) 실행
	docker run --rm -i --network host -v $$PWD:/work -w /work \
		postgres:16 psql "$(DB_URL)" -f internal/db/migrate/0001_init.sql
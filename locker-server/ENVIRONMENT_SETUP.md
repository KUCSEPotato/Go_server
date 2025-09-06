# 환경별 설정 가이드

## 개요
사물함 서버는 main(개발환경)과 prod(운영환경) 두 가지 환경으로 구성됩니다.

## Main 환경 (개발환경)

### 브랜치
- `main` 브랜치

### Docker 설정
**파일**: `docker-compose.yml`
```yaml
version: "3.9"
services:
  pg:
    image: postgres:16
    environment:
      POSTGRES_USER: locker
      POSTGRES_PASSWORD: locker
      POSTGRES_DB: locker
    ports: ["5432:5432"]
    volumes: [pgdata:/var/lib/postgresql/data]

  redis:
    image: redis:7
    ports: ["6379:6379"]

volumes:
  pgdata:
```

### 데이터베이스 접속 정보
- **Host**: localhost
- **Port**: 5432
- **Database**: locker
- **Username**: locker
- **Password**: locker

### Redis 접속 정보
- **Host**: localhost
- **Port**: 6379
- **Password**: 없음

### 서버 설정
- **URL**: http://localhost:3000
- **Swagger**: http://localhost:3000/swagger/index.html

### 실행 방법
```bash
# 1. 도커 컨테이너 실행
cd /Users/potato/Desktop/Dev/Go_server/locker-server
docker compose up -d

# 2. 서버 실행
go run ./cmd/server/main.go

# 3. 확인
docker ps
```

### 데이터베이스 접속
```bash
# PostgreSQL 컨테이너 접속
docker exec -it locker-server-pg-1 psql -U locker -d locker

# Redis 컨테이너 접속
docker exec -it locker-server-redis-1 redis-cli
```

### 테스트 설정
**파일**: `tests/load_test_config.py`
```python
# 서버 설정
BASE_URL = "http://localhost:3000"

# 데이터베이스 설정
# DB_HOST=localhost
# DB_PORT=5432
# DB_NAME=locker
# DB_USER=locker
# DB_PASSWORD=locker
```

---

## Prod 환경 (운영환경)

### 브랜치
- `prod` 브랜치

### Docker 설정
**파일**: `docker-compose.prod.yml`
```yaml
version: "3.9"
services:
  app:
    build: .
    ports:
      - "3000:3000"
    depends_on:
      - pg
      - redis
    environment:
      - DB_HOST=pg
      - DB_PORT=5432
      - DB_NAME=locker
      - DB_USER=locker
      - DB_PASSWORD=secure_password_2024
      - REDIS_HOST=redis
      - REDIS_PORT=6379

  pg:
    image: postgres:16
    environment:
      POSTGRES_USER: locker
      POSTGRES_PASSWORD: secure_password_2024
      POSTGRES_DB: locker
    ports:
      - "5432:5432"
    volumes:
      - pgdata:/var/lib/postgresql/data
      - ./internal/db/migrate:/docker-entrypoint-initdb.d

  redis:
    image: redis:7
    ports:
      - "6379:6379"

volumes:
  pgdata:
```

### 환경변수 설정
**파일**: `.env.prod`
```bash
DB_HOST=pg
DB_PORT=5432
DB_NAME=locker
DB_USER=locker
DB_PASSWORD=secure_password_2024
REDIS_HOST=redis
REDIS_PORT=6379
```

### 데이터베이스 접속 정보
- **Host**: pg (도커 내부) / 15.164.164.194 (외부)
- **Port**: 5432
- **Database**: locker
- **Username**: locker
- **Password**: secure_password_2024

### Redis 접속 정보
- **Host**: redis (도커 내부) / 15.164.164.194 (외부)
- **Port**: 6379
- **Password**: 없음

### 서버 설정
- **AWS EC2**: 15.164.164.194:3000
- **Swagger**: http://15.164.164.194:3000/swagger/index.html

### AWS EC2 배포
```bash
# 1. EC2 접속
ssh -i "locker-server-key.pem" ubuntu@15.164.164.194

# 2. 프로젝트 업데이트
cd Go_server/locker-server
git pull origin prod

# 3. 운영환경 실행
docker-compose -f docker-compose.prod.yml up -d --build

# 4. 확인
docker ps
curl http://localhost:3000/api/v1/health
```

### 데이터베이스 접속 (운영환경)
```bash
# PostgreSQL 컨테이너 접속
docker exec -it locker-server-pg-1 psql -U locker -d locker

# Redis 컨테이너 접속
docker exec -it locker-server-redis-1 redis-cli
```

---

## 데이터베이스 초기화

### 개발환경 (Main)
```sql
-- locker_assignments 초기화
TRUNCATE locker_assignments RESTART IDENTITY;

-- locker_info 초기화  
UPDATE locker_info SET owner = NULL;

-- 확인
SELECT * FROM locker_assignments;
SELECT * FROM locker_info;
```

### 운영환경 (Prod)
운영환경에서는 초기화 스크립트가 자동으로 실행됩니다:
- `internal/db/migrate/001_init.sql`

---

## 환경 전환

### Main에서 Prod로
```bash
# 1. 변경사항 커밋
git add .
git commit -m "development changes"

# 2. prod 브랜치로 전환
git checkout prod

# 3. main 변경사항 병합
git merge main

# 4. 운영환경 배포
git push origin prod
```

### Prod에서 Main으로
```bash
# 1. main 브랜치로 전환
git checkout main

# 2. 개발 계속
# (운영환경은 별도 유지)
```

---

## 부하 테스트 설정

### Main 환경 테스트
```bash
cd tests
python3 -m venv venv
source venv/bin/activate
pip install -r requirements.txt
python3 load_test.py
```

### Prod 환경 테스트
```bash
# 환경변수 설정
export BASE_URL="http://15.164.164.194:3000"
export DB_HOST="15.164.164.194"
export DB_PASSWORD="secure_password_2024"

# 테스트 실행
python3 load_test.py
```

---

## 주의사항

1. **비밀번호 관리**
   - Main: `locker` (개발용)
   - Prod: `secure_password_2024` (운영용)

2. **포트 충돌 방지**
   - 두 환경을 동시에 실행하지 않도록 주의

3. **데이터 백업**
   - 운영환경 데이터는 정기적으로 백업 필요

4. **브랜치 관리**
   - Main: 개발 전용
   - Prod: 운영 배포 전용

# 고려대학교 정보대학 제9대 학생회 사물함 서버 개발
- 기술 스택: Go + Fiber + postgresql

# 실행 방법
1. git clone
``` bash
git clone https://github.com/KUCSEPotato/Go_server.git
```
2. dir 이동
``` bash
cd locker-server
```
3. 서버 실행
``` bash
go run ./cmd/server/main.go
```
4. swagger 문서 확인
``` bash
http://localhost:3000/swagger/index.html
```
5. 서버 닫기
- ctrl + c
- ctrl + c가 안될 경우 -> ctrl + z
  - failed to listen: listen tcp4 :3000: bind: address already in use 인 경우
    - lsof -i :3000
    - kill -9 "pid"

# 구현 사항
- swag
  - swag init -g cmd/server/main.go -o docs/
- docker
  - docker 띄우기
    - main
      - docker compose -f docker/docker-compose.yml -p locker-dev up -d --build
    - docker compose up -d
    - prod
      - cd /Users/potato/Desktop/Dev/Go_server/locker-server && docker compose -f docker/docker-compose.prod.yml up --build
      - docker build -f docker/Dockerfile -t locker-server:prod .
      - docker compose -f docker/docker-compose.prod.yml up -d --build
      - docker compose -f docker/docker-compose.prod.yml down --remove-orphans
      - docker rm -f locker-prod-redis 2>/dev/null || echo "Redis container not found or already removed"

      - 충돌 날때
        # 1. 실행 중인 컨테이너들 강제 중지 및 삭제
          docker stop locker-prod-app locker-prod-pg locker-prod-redis
          docker rm locker-prod-app locker-prod-pg locker-prod-redis

          # 2. 네트워크도 삭제 (필요한 경우)
          docker network rm locker-prod-network

          # 3. 다시 컨테이너 시작
          docker compose -f docker/docker-compose.prod.yml up --build
  - 실행 확인
    - docker ps
    - docker ps -a
    - docker compose ps
    - docker info
  - postgresql 컨테이너 접속
    - [main] docker exec -it locker-dev-pg psql -U locker -d locker
    - [prod] docker exec -it locker-prod-pg psql -U locker -d locker
      ``` bash
      -- 모든 테이블 목록 확인
      \dt

      -- 각 테이블의 구조 확인
      \d users
      \d locker_info
      \d locker_assignments
      \d locker_locations
      \d auth_refresh_tokens

      -- 테이블 데이터 확인
      SELECT * FROM users;
      SELECT * FROM locker_info;
      SELECT * FROM locker_assignments;
      SELECT * FROM locker_locations;

      -- PostgreSQL 접속 종료
      \q
      
      ```
      - # locker_assignments 초기화
      - TRUNCATE locker_assignments RESTART IDENTITY;
      - # locker_info 초기화
      - UPDATE locker_info SET owner = NULL;
      - # 확인
      - SELECT * FROM locker_assignments;
      - SELECT * FROM locker_info;
    - docker exec -it locker-server-redis-1 redis-cli

  - 테스트 스크립트 이후 디비가 오염된 경우
    - local
      ``` bash
      docker exec locker-server-pg-1 psql -U locker -d locker -c "DELETE FROM locker_assignments; UPDATE locker_info SET owner = NULL;"
      docker exec locker-server-redis-1 redis-cli FLUSHALL
      ```
    - aws
      ``` bash
      docker exec locker-prod-pg psql -U locker -d locker -c "DELETE FROM locker_assignments; UPDATE locker_info SET owner = NULL;"
      docker exec locker-prod-redis redis-cli FLUSHALL
      ```
  - token 관련 데이터 초기화
    - local
      ``` bash
      # refresh token 테이블 비우기
      docker exec locker-prod-pg psql -U locker -d locker -c "TRUNCATE auth_refresh_tokens RESTART IDENTITY;"
      # Redis blacklist 비우기  
      docker exec locker-prod-redis redis-cli FLUSHALL
      ```
    - aws
      ``` bash
      # refresh token 테이블 비우기
      docker exec locker-prod-pg psql -U locker -d locker -c "TRUNCATE auth_refresh_tokens RESTART IDENTITY;"
      # Redis blacklist 비우기
      docker exec locker-prod-redis redis-cli FLUSHALL
      ```
- 서버 닫을 때는 컨트롤 + z || 컨트롤 + c (컨트롤 + z 사용시 아래의 명령어 사용 필요)
  - failed to listen: listen tcp4 :3000: bind: address already in use 인 경우
    - lsof -i :3000
    - kill -9 "pid"
- swag 명령어가 안될 경우
  - export PATH=$PATH:$(go env GOPATH)/bin 
  - 환경 변수 설정 필요
- ssh 접속
  - ssh -i "/Users/potato/Desktop/Dev/locker-server-key.pem" ubuntu@43.203.248.2

# 로컬 DB를 AWS로 마이그레이션
``` bash
# 1. 로컬에서 데이터베이스 덤프 생성
docker exec locker-prod-pg pg_dump -U locker -d locker --clean --if-exists > locker_db_dump.sql

# 2. 덤프 파일을 AWS 서버로 전송
scp -i "/Users/potato/Desktop/Dev/locker-server-key.pem" locker_db_dump.sql ubuntu@43.201.95.94:~/

# 3. AWS 서버에서 Docker 컨테이너 실행 후 데이터 복원
# AWS 서버에 SSH 접속 후:
# docker exec -i locker-prod-pg psql -U locker -d locker < locker_db_dump.sql

# 4. 복원 확인
# docker exec -it locker-prod-pg psql -U locker -d locker -c "\dt"
# docker exec -it locker-prod-pg psql -U locker -d locker -c "SELECT COUNT(*) FROM users;"
```

# AWS에 최신 데이터베이스 동기화
``` bash
# 방법 1: 로컬 DB 덤프를 AWS로 복사
# 1. 로컬에서 현재 DB 덤프 생성
docker exec locker-prod-pg pg_dump -U locker -d locker --clean --if-exists > locker_db_dump_new.sql

# 2. AWS 서버로 전송
scp -i "/Users/potato/Desktop/Dev/locker-server-key.pem" locker_db_dump_new.sql ubuntu@43.203.248.2:~/

# 3. AWS 서버에서 복원 (SSH 접속 후)
docker exec -i locker-prod-pg psql -U locker -d locker < locker_db_dump_new.sql

# 방법 2: AWS에서 직접 데이터 재생성
# 1. 기존 데이터 삭제
docker exec -it locker-prod-pg psql -U locker -d locker -c "
DELETE FROM locker_assignments;
DELETE FROM locker_info; 
DELETE FROM locker_locations;
"

# 2. 새로운 위치 및 사물함 데이터 생성
docker exec -it locker-prod-pg psql -U locker -d locker -c "
INSERT INTO locker_locations (location_id, name) VALUES
(1, '정보관 B1 엘리베이터 옆1'), (2, '정보관 B1 엘리베이터 옆2'),
(3, '정보관 B1 기계실 옆'), (4, '정보관 2층'), (5, '정보관 3층'),
(6, '과학도서관 6층 620호 옆'), (7, '과학도서관 6층 614호 옆');
SELECT setval('locker_locations_location_id_seq', 7, true);

INSERT INTO locker_info (locker_id, location_id) VALUES
(101, 1), (102, 1), (103, 1), (104, 2), (105, 2), (106, 2),
(201, 3), (202, 3), (301, 4), (302, 4), (303, 4), (304, 4),
(401, 5), (402, 5), (403, 5), (404, 5),
(501, 6), (502, 6), (503, 6), (601, 7), (602, 7), (603, 7);
"

# 3. 사용자 데이터 동기화
docker exec -it locker-prod-pg psql -U locker -d locker -c "
TRUNCATE auth_refresh_tokens RESTART IDENTITY;
TRUNCATE users RESTART IDENTITY CASCADE;
INSERT INTO users (student_id, name, phone_number, created_at, updated_at) VALUES 
('2023321234', '홍길동', '01012345678', NOW(), NOW()),
('2023325678', '김철수', '01087654321', NOW(), NOW());
"
```

# 홍길동, 김철수 학번 변경
``` bash
docker exec -it locker-prod-pg psql -U locker -d locker -c "INSERT INTO users (student_id, name, phone_number, created_at, updated_at) VALUES ('2023321234', '홍길동', '01012345679', NOW(), NOW());"

docker exec -it locker-prod-pg psql -U locker -d locker -c "INSERT INTO users (student_id, name, phone_number, created_at, updated_at) VALUES ('2023325678', '김철수', '01087654322', NOW(), NOW());"

docker exec -it locker-prod-pg psql -U locker -d locker -c "DELETE FROM users WHERE student_id IN ('20231234', '20231235');"

docker exec -it locker-prod-pg psql -U locker -d locker -c "UPDATE users SET phone_number = '01012345678' WHERE student_id = '2023321234';"

docker exec -it locker-prod-pg psql -U locker -d locker -c "UPDATE users SET phone_number = '01087654321' WHERE student_id = '2023325678';"

docker exec -it locker-prod-pg psql -U locker -d locker -c "SELECT * FROM users ORDER BY student_id;"
```

# user 테이블 초기화 및 재설정
``` bash
# 1. 외래키 제약 때문에 관련 테이블들도 함께 비워야 함
docker exec -it locker-prod-pg psql -U locker -d locker -c "TRUNCATE auth_refresh_tokens RESTART IDENTITY;"
docker exec -it locker-prod-pg psql -U locker -d locker -c "DELETE FROM locker_assignments;"
docker exec -it locker-prod-pg psql -U locker -d locker -c "UPDATE locker_info SET owner = NULL;"
docker exec -it locker-prod-pg psql -U locker -d locker -c "TRUNCATE users RESTART IDENTITY CASCADE;"

# 2. 테스트 사용자 재추가
docker exec -it locker-prod-pg psql -U locker -d locker -c "INSERT INTO users (student_id, name, phone_number, created_at, updated_at) VALUES ('2023321234', '홍길동', '01012345678', NOW(), NOW());"
docker exec -it locker-prod-pg psql -U locker -d locker -c "INSERT INTO users (student_id, name, phone_number, created_at, updated_at) VALUES ('2023325678', '김철수', '01087654321', NOW(), NOW());"

# 3. 확인
docker exec -it locker-prod-pg psql -U locker -d locker -c "SELECT * FROM users ORDER BY student_id;"

# 4. Redis 캐시도 비우기 (선택적)
docker exec locker-prod-redis redis-cli FLUSHALL
```

# AWS EC2 Docker 권한 설정 (최초 1회)
``` bash
# Docker 그룹에 현재 사용자 추가
sudo usermod -aG docker $USER

# 새로운 그룹 설정 적용 (재로그인 또는 아래 명령어)
newgrp docker

# 또는 재로그인
# exit 후 다시 ssh 접속

# Docker 서비스 시작
sudo systemctl start docker
sudo systemctl enable docker

# 권한 확인
docker ps
```

# test data setting
- student
  - {"student_id":"2023321234","name":"홍길동","phone_number":"01012345678"}
  - {"student_id":"2023325678","name":"김철수","phone_number":"01087654321"}
- curl test flow
  1. 로그인
    ``` bash
    curl -X POST http://localhost:3000/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"student_id":"2023321234","name":"홍길동","phone_number":"01012345678"}'
    ```
    ``` bash
    curl -X POST http://localhost:3000/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"student_id":"2023325678","name":"김철수","phone_number":"01087654321"}'
    ```
  2. 사물함 선점
    ``` bash
    curl -X POST http://localhost:3000/api/v1/lockers/101/hold \
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN"
    ```
  3. 사물함 확정
    ``` bash
    curl -X POST http://localhost:3000/api/v1/lockers/101/confirm \
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN"
    ```
  4. 사물함 목록 조회
  ``` bash
  curl -X GET http://localhost:3000/api/v1/lockers \
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN"
  ```
  5. 사물함 해제
  ``` bash
    curl -X POST http://localhost:3000/api/v1/lockers/101/release \
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN"
    ```
  6. 내 사물함 조회 [250904 여기까지 함]
  ``` bash
  curl -X GET http://localhost:3000/api/v1/lockers/me \
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN"
  ```
  7. jwt token refresh
  ``` bash
    curl -X POST http://localhost:3000/api/v1/auth/refresh \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer ACCESS_TOKEN" \
  -d '{"refresh_token":"REFRESH_TOKEN"}'
  ```
  8. health check
  ``` bash
  curl -X GET http://localhost:3000/api/v1/health \
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN"
  ```

## 필요한 구현 사항
1. 학번, 이름, 전화번호로 로그인
    - 학번으로 JWT token 발급
    - 학생회 홈페이지에 붙어있으나 기존 로그인은 단체(ex. 컴퓨터학과 학생회)이므로 새로운 테이블 개설 필요
        - 학번이 db의 primary key
2. 전체 사물함 정보 불러오기
    - 사물함 정보를 저장하는 테이블 필요
3. 사물함 신청
    - db 혹은 redis 사용하여 선착순 구현
    
## db 구조
https://dbdiagram.io/d/locker-sever-db-68b51f3a777b52b76c695e34
![alt text](/locker-server/image-1.png)
``` dbml
Project locker_reservation {
  database_type: 'PostgreSQL'
  Note: '사물함 선착순 예약 시스템'
}

/* ===== ENUM ===== */
Enum assignment_state {
  hold
  confirmed
  cancelled
  expired
}

/* ===== USERS ===== */
Table users {
  student_id      varchar(20)           [pk, not null, note: '학번 (PK)']
  name            varchar(100)  [not null]
  phone_number    varchar(32)   [not null, unique, note: '중복 가입 방지']
  // is_admin      boolean       [not null, default: false] // 필요 시 활성화
  created_at      timestamp     [not null, default: `now()`]
  updated_at      timestamp     [not null, default: `now()`]
}

/* ===== REFRESH TOKENS ===== */
Table auth_refresh_tokens {
  id              bigserial     [pk]
  student_id      int           [not null]
  token_hash      text          [not null, unique, note: '평문 저장 금지(해시만 저장)']
  issued_at       timestamp     [not null, default: `now()`, note: '발급 시각']
  expires_at      timestamp     [not null, note: '만료 시각']
  revoked_at      timestamp     [note: '로그아웃/강제 무효화 시각']
  user_agent      text          [note: '발급 당시 UA']
  ip              varchar(45)   [note: 'IPv4/IPv6 문자열(또는 inet)']

  Note: 'JWT 재발급/로그아웃(무효화) 관리'
}

Ref: auth_refresh_tokens.student_id > users.student_id

/* ===== LOCATIONS (코드값 테이블화) ===== */
Table locker_locations {
  location_id     serial        [pk]
  /* 
  정보관 지하 1층 - 엘리베이터: 1 
  정보관 지하 1층 - 기계실: 2 
  정보관 2층: 3 
  정보관 3층: 4 
  과학도서관 6층 - 왼쪽: 5 
  과학도서관 6층 - 오른쪽: 6 
  */
  name            text          [not null, unique, note: '예: 정보관 B1 엘리베이터, 정보관 B1 기계실, 정보관 2층, ...']
}

/* ===== LOCKERS (개별 사물함) ===== */
Table locker_info {
  locker_id       int           [pk, not null, note: '건물 코드 + 내부 코드 등 외부에서 부여한 고유값 가능']
  owner           int           [default: null, note: '선점한 사람의 학번(없으면 NULL)', unique] // 한 유저는 동시에 1개만 소유
  location_id     int           [not null, note: '사물함 위치 FK']

  Note: 'owner != NULL이면 선점 상태'
}

Ref: locker_info.owner > users.student_id
Ref: locker_info.location_id > locker_locations.location_id

/* ===== OVERALL INFO (권장: VIEW, 여기선 설명용 테이블) ===== */
Table overall_locker_info {
  remaining       int           [note: '권장: VIEW로 계산. 여기선 설명용 placeholder']
  // 필요 시 location별 집계 컬럼 추가 가능
}

/* ===== ASSIGNMENTS (선점/확정 히스토리) ===== */
Table locker_assignments {
  assignment_id   bigserial         [pk]
  locker_id       int               [not null]
  student_id      int               [not null]
  state           assignment_state  [not null]
  hold_expires_at timestamp
  confirmed_at    timestamp
  released_at     timestamp
  created_at      timestamp         [not null, default: `now()`]

  Note: '활성 상태(hold/confirmed) 동시 1건 제한은 부분 유니크 인덱스로 마이그레이션에서 구현'
}

Ref: locker_assignments.student_id > users.student_id
Ref: locker_assignments.locker_id  > locker_info.locker_id

/* ===== 인덱스 & 제약 =====
-- PostgreSQL 마이그레이션에서 권장:
-- 1) 사물함별 활성 1건 제한 (hold/confirmed)
--   CREATE UNIQUE INDEX ux_active_assignment_per_locker
--   ON locker_assignments (locker_id)
--   WHERE state IN ('hold','confirmed');

-- 2) 사용자별 활성 1건 제한 (hold/confirmed)
--   CREATE UNIQUE INDEX ux_active_assignment_per_user
--   ON locker_assignments (student_id)
--   WHERE state IN ('hold','confirmed');

-- 3) 조회 보조
--   CREATE INDEX idx_assignments_lookup ON locker_assignments (locker_id, state);

-- 4) overall_locker_info는 VIEW 권장:
--   CREATE VIEW v_overall_locker_info AS
--   SELECT COUNT(*) FILTER (WHERE NOT EXISTS (
--     SELECT 1 FROM locker_assignments a
--     WHERE a.locker_id = l.locker_id
--       AND a.state IN ('confirmed','hold')
--       AND (a.state = 'confirmed'
--         OR (a.state = 'hold' AND (a.hold_expires_at IS NULL OR a.hold_expires_at > now()))
--       )
--   )) AS remaining
--   FROM locker_info l;
*/

```
## 프론트 화면
![alt text](/locker-server/image.png)
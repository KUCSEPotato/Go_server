# TODO
1. swagger docs 제작 [done 250902]
2. docker로 서버 띄어서 테스트 [done 250902]
3. api test
  - docker에 db 띄우기
4. api 수정


# 고려대학교 정보대학 제9대 학생회 사물함 서버 개발
- 기술 스택: Go + Fiber + postgresql

# 구현 사항
- docker
  - docker 띄우기
    - docker compose up -d
  - 실행 확인
    - docker ps
    - docker ps -a
    - docker compose ps
    - docker info
  - postgresql 컨테이너 접속
    - docker exec -it locker-server-pg-1 psql -U locker -d locker
      - # locker_assignments 초기화
      - TRUNCATE locker_assignments RESTART IDENTITY;
      - # locker_info 초기화
      - UPDATE locker_info SET owner = NULL;
      - # 확인
      - SELECT * FROM locker_assignments;
      - SELECT * FROM locker_info;
    - docker exec -it locker-server-redis-1 redis-cli
- 서버 닫을 때는 컨트롤 + z
  - failed to listen: listen tcp4 :3000: bind: address already in use 인 경우
    - lsof -i :3000
    - kill -9 "pid"
- swag 명령어가 안될 경우
  - export PATH=$PATH:$(go env GOPATH)/bin 
  - 환경 변수 설정 필요

# test data setting
- student
  - {"student_id":"20231234","name":"홍길동","phone_number":"01012345678"}
  - {"student_id":"20231235","name":"김철수","phone_number":"01087654321"}
- curl test flow
  1. 로그인
    ``` bash
    curl -X POST http://localhost:3000/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"student_id":"20231234","name":"홍길동","phone_number":"01012345678"}'
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
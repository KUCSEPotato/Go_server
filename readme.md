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
- 서버 닫을 때는 컨트롤 + z
  - failed to listen: listen tcp4 :3000: bind: address already in use 인 경우
    - lsof -i :3000
    - kill -9 "pid"
- swag 명령어가 안될 경우
  - export PATH=$PATH:$(go env GOPATH)/bin 
  - 환경 변수 설정 필요

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
  student_id      int           [pk, not null, note: '학번 (PK)']
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

## project 디렉토리 구조
``` text
locker-server/
├─ cmd/                                   # 여러 실행 바이너리의 엔트리포인트 모음 (CLI, server 등)
│  └─ server/
│     └─ main.go                          # 서버 시작점: Fiber 부트, 미들웨어 장착, 라우터 등록, Listen
│
├─ internal/                              # 외부에서 import 불가(Go의 internal 규칙) — 앱의 내부 구현
│  ├─ api/                                # HTTP 레이어(라우터/핸들러/미들웨어/DTO)
│  │  ├─ router.go                        # 라우트 트리 구성: /api, /v1 그룹핑 및 핸들러 바인딩
│  │  ├─ middleware/                      # 요청 전/후 공통 처리 (auth, logger 확장, rate limit 등)
│  │  │  ├─ auth.go                       # JWT 검증, 권한 체크, RequestID 등
│  │  │  └─ recover.go                    # panic 복구(필요 시 커스터마이즈)
│  │  ├─ handlers/                        # 실제 엔드포인트 로직 (HTTP 입출력에 집중)
│  │  │  ├─ auth.go                       # 로그인, 토큰 재발급/무효화
│  │  │  └─ locker.go                     # 사물함 조회/선점(Hold)/확정(Confirm)/해제(Release)
│  │  └─ dto/                             # 요청/응답 바디 스키마(입력 검증, 응답 포맷 정의)
│  │     ├─ auth_dto.go                   # 로그인/리프레시 DTO
│  │     └─ locker_dto.go                 # 사물함 관련 DTO
│  │
│  ├─ service/                            # 유스케이스 계층: 비즈니스 플로우 오케스트레이션
│  │  ├─ auth_service.go                  # 로그인→토큰 발급, 리프레시 검증 로직
│  │  └─ locker_service.go                # Redis 홀드→DB 기록→확정 트랜잭션 등 핵심 시나리오
│  │
│  ├─ domain/                             # 도메인 모델(엔터티/값객체)과 비즈니스 규칙
│  │  ├─ users.go                         # User 엔터티, 도메인 규칙/에러
│  │  └─ lockers.go                       # Locker/Assignment 엔터티, 상태 전이 규칙
│  │
│  ├─ repository/                         # DB 접근 계층(쿼리/트랜잭션) — pgx, SQL 빌더, 캐시 등
│  │  ├─ user_repo.go                     # users SELECT/INSERT/UPDATE 등
│  │  ├─ locker_repo.go                   # locker_info/locations CRUD
│  │  └─ assignment_repo.go               # locker_assignments 히스토리/유니크 제약 충돌 처리
│  │
│  ├─ db/                                 # 인프라(DB) 연결/마이그레이션 관련
│  │  ├─ postgres.go                      # pgxpool 초기화, 헬스체크, 연결 파라미터 설정
│  │  └─ migrate/                         # SQL 마이그레이션 파일들(버전 관리)
│  │     └─ 0001_init.sql                 # 초기 스키마: 테이블/인덱스/enum 생성
│  │
│  ├─ cache/                              # Redis 등 캐시/분산락/레이트리밋 스토리지
│  │  └─ redis.go                         # Redis 클라이언트 초기화(SetNX, TTL, Lua 등 유틸)
│  │
│  └─ util/                               # 범용 유틸(순수 로직): JWT, 해시, 응답 헬퍼, 시간/문자열 등
│     ├─ jwt.go                           # JWS 기반 Access 토큰 발급/검증
│     ├─ hash.go                          # 해시/난수 토큰 생성(Refresh 해시용)
│     └─ response.go                      # 공통 응답 포맷/에러 변환
│
├─ configs/                               # 설정 파일(.env 샘플, YAML/JSON 설정 등)
│  └─ config.example.env                  # 환경변수 예시(포트, DB_URL, JWT 시크릿/TTL 등)
│
├─ scripts/                               # 개발/운영 스크립트(로컬 부트스트랩, 데이터 시드, 린트 등)
│  ├─ dev_up.sh                           # 로컬 개발용 부트스트랩 스크립트
│  └─ seed.sql                            # 샘플 데이터 insert (선택)
│
├─ test/                                  # 통합/엔드투엔드 테스트, 테스트 픽스처
│  └─ e2e/                                # http 테스트, 시나리오별 케이스
│
├─ docker-compose.yml                     # 로컬 infra(Postgres/Redis) 기동
├─ Makefile                               # 자주 쓰는 명령어 단축(run, migrate, test 등)
├─ go.mod                                 # Go 모듈 정의
├─ go.sum                                 # 의존성 해시
├─ .gitignore                             # 바이너리/캐시/환경파일 무시
└─ README.md                              # 프로젝트 개요, 실행 방법, API 문서 링크
```
-- 데이터 정리 및 마이그레이션 스크립트

-- 1. users, locker_assignments, auth_refresh_tokens 데이터 삭제
TRUNCATE TABLE users CASCADE;
TRUNCATE TABLE locker_assignments CASCADE;
TRUNCATE TABLE auth_refresh_tokens CASCADE;

-- 2. locker_info 데이터 유지 (필요한 경우 추가 작업)
-- 이 테이블은 데이터를 유지하므로 변경 없음

-- 3. Redis 데이터 초기화
-- Redis에 저장된 모든 키를 삭제하여 초기화합니다
-- 명령어 예시:
-- docker exec locker-prod-redis redis-cli FLUSHALL

-- 4. 데이터베이스 스키마 덤프를 AWS로 마이그레이션
-- pg_dump 명령어를 사용하여 스키마와 데이터를 AWS로 업로드
-- 예시:
-- pg_dump -U locker -h localhost -d locker | psql -h <AWS_RDS_ENDPOINT> -U <AWS_USER> -d <AWS_DB_NAME>

-- 5. 마이그레이션 후 확인
-- AWS에서 데이터가 정상적으로 마이그레이션되었는지 확인
SELECT * FROM locker_info;

-- 6. 추가적인 데이터 검증
-- 필요한 경우 데이터 검증을 위한 쿼리를 추가합니다
-- 예시:
-- SELECT COUNT(*) FROM locker_assignments WHERE state='confirmed';

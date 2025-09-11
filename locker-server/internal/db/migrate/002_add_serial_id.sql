-- 사용자 테이블에 serial_id 추가 및 Primary Key 변경

-- 1. 새로운 serial_id 컬럼 추가
ALTER TABLE users ADD COLUMN id SERIAL;

-- 2. 기존 제약조건들 임시 제거
ALTER TABLE auth_refresh_tokens DROP CONSTRAINT IF EXISTS fk_refresh_student;
ALTER TABLE locker_info DROP CONSTRAINT IF EXISTS fk_locker_owner;
ALTER TABLE locker_assignments DROP CONSTRAINT IF EXISTS fk_assignment_student;

-- 3. 기존 Primary Key 제거 및 새로운 Primary Key 설정
ALTER TABLE users DROP CONSTRAINT IF EXISTS users_pkey;
ALTER TABLE users ADD PRIMARY KEY (id);

-- 4. student_id에서 UNIQUE 제약조건 제거 (중복 허용)
ALTER TABLE users DROP CONSTRAINT IF EXISTS users_student_id_key;

-- 5. phone_number UNIQUE 제약조건도 제거 (필요시)
ALTER TABLE users DROP CONSTRAINT IF EXISTS users_phone_number_key;

-- 6. 외래키 제약조건 다시 추가 (student_id 기반 유지)
ALTER TABLE auth_refresh_tokens 
ADD CONSTRAINT fk_refresh_student 
FOREIGN KEY (student_id) REFERENCES users(student_id);

ALTER TABLE locker_info 
ADD CONSTRAINT fk_locker_owner 
FOREIGN KEY (owner) REFERENCES users(student_id);

ALTER TABLE locker_assignments 
ADD CONSTRAINT fk_assignment_student 
FOREIGN KEY (student_id) REFERENCES users(student_id);

-- 7. student_id에 인덱스 추가 (성능을 위해)
CREATE INDEX IF NOT EXISTS idx_users_student_id ON users(student_id);

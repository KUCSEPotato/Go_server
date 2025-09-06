-- 사물함 예약 시스템 데이터베이스 스키마 (로컬 DB와 동일)

-- ENUM 타입 생성
CREATE TYPE assignment_state AS ENUM ('hold', 'confirmed', 'cancelled', 'expired');

-- Users 테이블
CREATE TABLE IF NOT EXISTS users (
    student_id VARCHAR(20) PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    phone_number VARCHAR(32) NOT NULL UNIQUE,
    created_at TIMESTAMP WITHOUT TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITHOUT TIME ZONE NOT NULL DEFAULT NOW()
);

-- Locker Locations 테이블
CREATE TABLE IF NOT EXISTS locker_locations (
    location_id SERIAL PRIMARY KEY,
    name TEXT NOT NULL UNIQUE
);

-- Locker Info 테이블
CREATE TABLE IF NOT EXISTS locker_info (
    locker_id INTEGER PRIMARY KEY,
    owner VARCHAR(20) DEFAULT NULL,
    location_id INTEGER NOT NULL,
    CONSTRAINT fk_locker_location FOREIGN KEY (location_id) REFERENCES locker_locations(location_id),
    CONSTRAINT fk_locker_owner FOREIGN KEY (owner) REFERENCES users(student_id),
    CONSTRAINT locker_info_owner_key UNIQUE (owner)
);

-- Auth Refresh Tokens 테이블
CREATE TABLE IF NOT EXISTS auth_refresh_tokens (
    id SERIAL PRIMARY KEY,
    student_id VARCHAR(20) NOT NULL,
    token_hash TEXT NOT NULL UNIQUE,
    issued_at TIMESTAMP WITHOUT TIME ZONE NOT NULL DEFAULT NOW(),
    expires_at TIMESTAMP WITHOUT TIME ZONE NOT NULL,
    revoked_at TIMESTAMP WITHOUT TIME ZONE,
    user_agent TEXT,
    ip VARCHAR(45),
    CONSTRAINT fk_refresh_student FOREIGN KEY (student_id) REFERENCES users(student_id)
);

-- Locker Assignments 테이블
CREATE TABLE IF NOT EXISTS locker_assignments (
    assignment_id BIGSERIAL PRIMARY KEY,
    locker_id INTEGER NOT NULL,
    student_id VARCHAR(20) NOT NULL,
    state assignment_state NOT NULL,
    hold_expires_at TIMESTAMP WITHOUT TIME ZONE,
    confirmed_at TIMESTAMP WITHOUT TIME ZONE,
    released_at TIMESTAMP WITHOUT TIME ZONE,
    created_at TIMESTAMP WITHOUT TIME ZONE NOT NULL DEFAULT NOW(),
    CONSTRAINT fk_assignment_locker FOREIGN KEY (locker_id) REFERENCES locker_info(locker_id),
    CONSTRAINT fk_assignment_student FOREIGN KEY (student_id) REFERENCES users(student_id)
);

-- 인덱스 생성
CREATE INDEX IF NOT EXISTS idx_assignments_lookup ON locker_assignments (locker_id, state);
CREATE UNIQUE INDEX IF NOT EXISTS ux_active_assignment_per_locker ON locker_assignments (locker_id) 
WHERE state = ANY (ARRAY['hold'::assignment_state, 'confirmed'::assignment_state]);
CREATE UNIQUE INDEX IF NOT EXISTS ux_active_assignment_per_user ON locker_assignments (student_id) 
WHERE state = ANY (ARRAY['hold'::assignment_state, 'confirmed'::assignment_state]);

-- 기본 데이터 삽입
-- Locker Locations 데이터
INSERT INTO locker_locations (location_id, name) VALUES
(1, '정보관 B1 엘리베이터'),
(2, '정보관 B1 기계실'),
(3, '정보관 2층'),
(4, '정보관 3층'),
(5, '과학도서관 6층 왼쪽'),
(6, '과학도서관 6층 오른쪽')
ON CONFLICT (location_id) DO NOTHING;

-- 시퀀스 업데이트
SELECT setval('locker_locations_location_id_seq', 6, true);

-- Locker Info 데이터
INSERT INTO locker_info (locker_id, location_id) VALUES
(101, 1), (102, 1), (103, 1),
(201, 2), (202, 2),
(301, 3), (302, 3), (303, 3),
(401, 4), (402, 4),
(501, 5), (502, 5),
(601, 6), (602, 6)
ON CONFLICT (locker_id) DO NOTHING;

-- 테스트 사용자 추가
INSERT INTO users (student_id, name, phone_number) 
VALUES ('20231234', '홍길동', '01012345678')
ON CONFLICT (student_id) DO NOTHING;

-- 기본 사물함 데이터 삽입 (101~150번)
INSERT INTO locker_info (locker_id, location) 
VALUES 
    (101, 'Building A Floor 1'),
    (102, 'Building A Floor 1'),
    (103, 'Building A Floor 1'),
    (104, 'Building A Floor 1'),
    (105, 'Building A Floor 1'),
    (106, 'Building A Floor 1'),
    (107, 'Building A Floor 1'),
    (108, 'Building A Floor 1'),
    (109, 'Building A Floor 1'),
    (110, 'Building A Floor 1'),
    (111, 'Building A Floor 2'),
    (112, 'Building A Floor 2'),
    (113, 'Building A Floor 2'),
    (114, 'Building A Floor 2'),
    (115, 'Building A Floor 2'),
    (116, 'Building A Floor 2'),
    (117, 'Building A Floor 2'),
    (118, 'Building A Floor 2'),
    (119, 'Building A Floor 2'),
    (120, 'Building A Floor 2'),
    (121, 'Building B Floor 1'),
    (122, 'Building B Floor 1'),
    (123, 'Building B Floor 1'),
    (124, 'Building B Floor 1'),
    (125, 'Building B Floor 1'),
    (126, 'Building B Floor 1'),
    (127, 'Building B Floor 1'),
    (128, 'Building B Floor 1'),
    (129, 'Building B Floor 1'),
    (130, 'Building B Floor 1'),
    (131, 'Building B Floor 2'),
    (132, 'Building B Floor 2'),
    (133, 'Building B Floor 2'),
    (134, 'Building B Floor 2'),
    (135, 'Building B Floor 2'),
    (136, 'Building B Floor 2'),
    (137, 'Building B Floor 2'),
    (138, 'Building B Floor 2'),
    (139, 'Building B Floor 2'),
    (140, 'Building B Floor 2'),
    (141, 'Building C Floor 1'),
    (142, 'Building C Floor 1'),
    (143, 'Building C Floor 1'),
    (144, 'Building C Floor 1'),
    (145, 'Building C Floor 1'),
    (146, 'Building C Floor 1'),
    (147, 'Building C Floor 1'),
    (148, 'Building C Floor 1'),
    (149, 'Building C Floor 1'),
    (150, 'Building C Floor 1')
ON CONFLICT (locker_id) DO NOTHING;

-- 테스트 사용자 추가
INSERT INTO users (student_id, name, phone_number) 
VALUES ('20231234', '홍길동', '01012345678')
ON CONFLICT (student_id) DO NOTHING;

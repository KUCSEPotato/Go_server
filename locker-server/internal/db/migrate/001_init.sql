-- 사물함 예약 시스템 데이터베이스 스키마

-- Users 테이블
CREATE TABLE IF NOT EXISTS users (
    student_id VARCHAR(20) PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    phone_number VARCHAR(20) NOT NULL,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- Locker Info 테이블
CREATE TABLE IF NOT EXISTS locker_info (
    locker_id INTEGER PRIMARY KEY,
    location VARCHAR(100) NOT NULL,
    owner VARCHAR(20) REFERENCES users(student_id) ON DELETE SET NULL,
    status VARCHAR(20) DEFAULT 'available', -- available, held, occupied
    held_at TIMESTAMP,
    confirmed_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- Refresh Token 테이블
CREATE TABLE IF NOT EXISTS auth_refresh_tokens (
    token_id VARCHAR(255) PRIMARY KEY,
    student_id VARCHAR(20) NOT NULL REFERENCES users(student_id) ON DELETE CASCADE,
    expires_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP DEFAULT NOW()
);

-- Locker Assignments 테이블 (예약 이력)
CREATE TABLE IF NOT EXISTS locker_assignments (
    assignment_id SERIAL PRIMARY KEY,
    student_id VARCHAR(20) NOT NULL REFERENCES users(student_id) ON DELETE CASCADE,
    locker_id INTEGER NOT NULL REFERENCES locker_info(locker_id) ON DELETE CASCADE,
    assigned_at TIMESTAMP DEFAULT NOW(),
    status VARCHAR(20) DEFAULT 'active' -- active, released, expired
);

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

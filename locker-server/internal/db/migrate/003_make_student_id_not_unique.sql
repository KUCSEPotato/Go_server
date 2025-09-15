-- This migration changes user identification from student_id to serial_id
-- to allow for non-unique student_id.

BEGIN;

-- Step 1: Drop foreign key constraints that depend on users.student_id
ALTER TABLE auth_refresh_tokens DROP CONSTRAINT IF EXISTS fk_refresh_student;
ALTER TABLE locker_info DROP CONSTRAINT IF EXISTS fk_locker_owner;
ALTER TABLE locker_assignments DROP CONSTRAINT IF EXISTS fk_assignment_student;

-- Step 2: Drop unique constraint on users.student_id
ALTER TABLE users DROP CONSTRAINT IF EXISTS users_student_id_key;


-- Step 3: Modify auth_refresh_tokens table
ALTER TABLE auth_refresh_tokens ADD COLUMN user_serial_id BIGINT;

-- Populate user_serial_id from users table
UPDATE auth_refresh_tokens art
SET user_serial_id = u.serial_id
FROM users u
WHERE art.student_id = u.student_id;

ALTER TABLE auth_refresh_tokens ALTER COLUMN user_serial_id SET NOT NULL;
ALTER TABLE auth_refresh_tokens DROP COLUMN student_id;
ALTER TABLE auth_refresh_tokens ADD CONSTRAINT fk_refresh_user_serial FOREIGN KEY (user_serial_id) REFERENCES users(serial_id) ON DELETE CASCADE;
CREATE INDEX IF NOT EXISTS idx_auth_refresh_tokens_user_serial_id ON auth_refresh_tokens(user_serial_id);


-- Step 4: Modify locker_info table
ALTER TABLE locker_info DROP CONSTRAINT IF EXISTS locker_info_owner_key;
ALTER TABLE locker_info RENAME COLUMN owner TO owner_student_id;
ALTER TABLE locker_info ADD COLUMN owner_serial_id BIGINT;

-- Populate owner_serial_id from users table
UPDATE locker_info li
SET owner_serial_id = u.serial_id
FROM users u
WHERE li.owner_student_id = u.student_id;

ALTER TABLE locker_info ADD CONSTRAINT fk_locker_owner_serial FOREIGN KEY (owner_serial_id) REFERENCES users(serial_id) ON DELETE SET NULL;
ALTER TABLE locker_info ADD CONSTRAINT locker_info_owner_serial_id_key UNIQUE (owner_serial_id);
-- We probably don't need owner_student_id anymore
-- ALTER TABLE locker_info DROP COLUMN owner_student_id;


-- Step 5: Modify locker_assignments table
DROP INDEX IF EXISTS ux_active_assignment_per_user;
ALTER TABLE locker_assignments ADD COLUMN user_serial_id BIGINT;

-- Populate user_serial_id from users table
UPDATE locker_assignments la
SET user_serial_id = u.serial_id
FROM users u
WHERE la.student_id = u.student_id;

ALTER TABLE locker_assignments ALTER COLUMN user_serial_id SET NOT NULL;
ALTER TABLE locker_assignments DROP COLUMN student_id;
ALTER TABLE locker_assignments ADD CONSTRAINT fk_assignment_user_serial FOREIGN KEY (user_serial_id) REFERENCES users(serial_id) ON DELETE CASCADE;

CREATE UNIQUE INDEX IF NOT EXISTS ux_active_assignment_per_user ON locker_assignments (user_serial_id)
WHERE state = ANY (ARRAY['hold'::assignment_state, 'confirmed'::assignment_state]);


COMMIT;

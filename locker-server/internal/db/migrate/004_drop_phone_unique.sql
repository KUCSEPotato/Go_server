-- Allow duplicate phone numbers by dropping the unique constraint
-- Safe to run multiple times due to IF EXISTS
BEGIN;

ALTER TABLE users
  DROP CONSTRAINT IF EXISTS users_phone_number_key;

COMMIT;



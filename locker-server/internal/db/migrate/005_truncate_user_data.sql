-- This script truncates the users and auth_refresh_tokens tables to clear all user data.
-- It's used to reset user information, for example after changing the serial_id generation logic.

-- Truncate both tables, reset sequences, and handle foreign key constraints.
TRUNCATE TABLE users, auth_refresh_tokens RESTART IDENTITY CASCADE;

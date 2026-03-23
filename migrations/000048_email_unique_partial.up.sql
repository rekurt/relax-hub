-- Replace unconditional UNIQUE constraint on email with a partial one
-- that excludes empty emails (phone-only registrations).
ALTER TABLE users DROP CONSTRAINT IF EXISTS users_email_key;
DROP INDEX IF EXISTS idx_users_email;
CREATE UNIQUE INDEX idx_users_email_unique ON users (email) WHERE email != '';

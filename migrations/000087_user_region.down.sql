-- Remove region field from users
DROP INDEX IF EXISTS idx_users_region;
ALTER TABLE users DROP COLUMN IF EXISTS region;

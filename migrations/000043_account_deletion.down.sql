DROP INDEX IF EXISTS idx_users_deletion_scheduled;
ALTER TABLE users DROP COLUMN IF EXISTS deletion_scheduled_at;
ALTER TABLE users DROP COLUMN IF EXISTS deletion_requested_at;

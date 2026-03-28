DROP INDEX IF EXISTS idx_users_admin_sub_role;
ALTER TABLE users DROP COLUMN IF EXISTS admin_sub_role;

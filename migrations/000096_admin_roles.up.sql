-- Add admin_sub_role column to users table
ALTER TABLE users ADD COLUMN admin_sub_role VARCHAR(20) NOT NULL DEFAULT '';

-- Set existing admin users to super_admin
UPDATE users SET admin_sub_role = 'super_admin' WHERE role = 'admin';

-- Create index for admin sub-role lookups
CREATE INDEX idx_users_admin_sub_role ON users (admin_sub_role) WHERE admin_sub_role != '';

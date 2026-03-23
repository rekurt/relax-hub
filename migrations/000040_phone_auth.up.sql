-- Add phone_verified column to users table
ALTER TABLE users ADD COLUMN phone_verified BOOLEAN NOT NULL DEFAULT false;

-- Add unique index on phone where phone is not empty
CREATE UNIQUE INDEX idx_users_phone_unique ON users (phone) WHERE phone != '';

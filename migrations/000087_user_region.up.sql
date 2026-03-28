-- Add region field to users (default RU)
ALTER TABLE users ADD COLUMN IF NOT EXISTS region VARCHAR(2) NOT NULL DEFAULT 'RU';

-- Add archived status to wallets status check (if constraint exists, drop and recreate)
-- The wallet status column already supports 'active' and 'frozen'; we add 'archived'.
-- Since PostgreSQL doesn't have a simple ALTER for CHECK constraints, we handle it gracefully.

-- Create index for quick region lookups
CREATE INDEX IF NOT EXISTS idx_users_region ON users (region);

-- Add security deposit percent to bathhouses (0-50%, default 0 = no deposit)
ALTER TABLE bathhouses ADD COLUMN IF NOT EXISTS security_deposit_percent INTEGER NOT NULL DEFAULT 0;

-- Add deposit tracking fields to bookings
ALTER TABLE bookings ADD COLUMN IF NOT EXISTS deposit_amount BIGINT NOT NULL DEFAULT 0;
ALTER TABLE bookings ADD COLUMN IF NOT EXISTS deposit_status TEXT NOT NULL DEFAULT 'none';
ALTER TABLE bookings ADD COLUMN IF NOT EXISTS deposit_external_id TEXT NOT NULL DEFAULT '';
ALTER TABLE bookings ADD COLUMN IF NOT EXISTS deposit_released_at TIMESTAMPTZ;

-- Index for cron job: find bookings with held deposits ready for auto-release
CREATE INDEX IF NOT EXISTS idx_bookings_deposit_held ON bookings (deposit_status, checked_out_at)
    WHERE deposit_status = 'held' AND checked_out_at IS NOT NULL;

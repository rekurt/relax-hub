DROP INDEX IF EXISTS idx_bookings_deposit_held;
ALTER TABLE bookings DROP COLUMN IF EXISTS deposit_released_at;
ALTER TABLE bookings DROP COLUMN IF EXISTS deposit_external_id;
ALTER TABLE bookings DROP COLUMN IF EXISTS deposit_status;
ALTER TABLE bookings DROP COLUMN IF EXISTS deposit_amount;
ALTER TABLE bathhouses DROP COLUMN IF EXISTS security_deposit_percent;

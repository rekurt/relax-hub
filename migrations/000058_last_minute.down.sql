ALTER TABLE bookings DROP COLUMN IF EXISTS last_minute_discount;

ALTER TABLE bathhouses
    DROP CONSTRAINT IF EXISTS chk_last_minute_discount_percent,
    DROP CONSTRAINT IF EXISTS chk_last_minute_hours_threshold;

ALTER TABLE bathhouses
    DROP COLUMN IF EXISTS last_minute_enabled,
    DROP COLUMN IF EXISTS last_minute_discount_percent,
    DROP COLUMN IF EXISTS last_minute_hours_threshold;

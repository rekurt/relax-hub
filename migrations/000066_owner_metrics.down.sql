ALTER TABLE bookings DROP COLUMN IF EXISTS cancelled_by_owner;
ALTER TABLE bathhouses DROP COLUMN IF EXISTS avg_response_time_minutes;
ALTER TABLE bathhouses DROP COLUMN IF EXISTS response_rate;

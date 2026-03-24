ALTER TABLE bookings
    DROP COLUMN IF EXISTS hold_id,
    DROP COLUMN IF EXISTS rejection_reason;

ALTER TABLE bathhouses
    DROP CONSTRAINT IF EXISTS chk_booking_mode,
    DROP CONSTRAINT IF EXISTS chk_request_timeout,
    DROP COLUMN IF EXISTS booking_mode,
    DROP COLUMN IF EXISTS request_timeout;

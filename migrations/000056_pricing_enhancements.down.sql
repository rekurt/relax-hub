ALTER TABLE bathhouses
    DROP CONSTRAINT IF EXISTS chk_long_session_discount_percent,
    DROP CONSTRAINT IF EXISTS chk_long_session_threshold_hours,
    DROP CONSTRAINT IF EXISTS chk_base_capacity,
    DROP CONSTRAINT IF EXISTS chk_extra_guest_surcharge;

ALTER TABLE bathhouses
    DROP COLUMN IF EXISTS long_session_threshold_hours,
    DROP COLUMN IF EXISTS long_session_discount_percent,
    DROP COLUMN IF EXISTS base_capacity,
    DROP COLUMN IF EXISTS extra_guest_surcharge;

ALTER TABLE bookings
    DROP COLUMN IF EXISTS base_price,
    DROP COLUMN IF EXISTS long_session_discount,
    DROP COLUMN IF EXISTS extra_guest_surcharge;

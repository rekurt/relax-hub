ALTER TABLE bathhouses
    ADD COLUMN long_session_threshold_hours INT NOT NULL DEFAULT 4,
    ADD COLUMN long_session_discount_percent INT NOT NULL DEFAULT 0,
    ADD COLUMN base_capacity INT,
    ADD COLUMN extra_guest_surcharge BIGINT NOT NULL DEFAULT 0;

-- Set base_capacity to max_guests for existing rows
UPDATE bathhouses SET base_capacity = max_guests WHERE base_capacity IS NULL;

-- Now make it NOT NULL
ALTER TABLE bathhouses ALTER COLUMN base_capacity SET NOT NULL;

-- Add constraints
ALTER TABLE bathhouses
    ADD CONSTRAINT chk_long_session_discount_percent CHECK (long_session_discount_percent BETWEEN 0 AND 50),
    ADD CONSTRAINT chk_long_session_threshold_hours CHECK (long_session_threshold_hours BETWEEN 1 AND 12),
    ADD CONSTRAINT chk_base_capacity CHECK (base_capacity >= 1),
    ADD CONSTRAINT chk_extra_guest_surcharge CHECK (extra_guest_surcharge >= 0);

-- Add price breakdown columns to bookings table
ALTER TABLE bookings
    ADD COLUMN base_price BIGINT NOT NULL DEFAULT 0,
    ADD COLUMN long_session_discount BIGINT NOT NULL DEFAULT 0,
    ADD COLUMN extra_guest_surcharge BIGINT NOT NULL DEFAULT 0;

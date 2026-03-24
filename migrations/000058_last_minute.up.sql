ALTER TABLE bathhouses
    ADD COLUMN last_minute_enabled BOOLEAN NOT NULL DEFAULT false,
    ADD COLUMN last_minute_discount_percent INT NOT NULL DEFAULT 20,
    ADD COLUMN last_minute_hours_threshold INT NOT NULL DEFAULT 6;

ALTER TABLE bathhouses
    ADD CONSTRAINT chk_last_minute_discount_percent CHECK (last_minute_discount_percent BETWEEN 5 AND 50),
    ADD CONSTRAINT chk_last_minute_hours_threshold CHECK (last_minute_hours_threshold BETWEEN 2 AND 24);

ALTER TABLE bookings
    ADD COLUMN last_minute_discount BIGINT NOT NULL DEFAULT 0;

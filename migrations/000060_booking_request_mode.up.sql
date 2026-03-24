-- Add booking mode (instant vs request-based) to bathhouses
ALTER TABLE bathhouses
    ADD COLUMN booking_mode VARCHAR(10) NOT NULL DEFAULT 'instant',
    ADD COLUMN request_timeout INT NOT NULL DEFAULT 24;

ALTER TABLE bathhouses
    ADD CONSTRAINT chk_booking_mode CHECK (booking_mode IN ('instant', 'request')),
    ADD CONSTRAINT chk_request_timeout CHECK (request_timeout BETWEEN 1 AND 72);

-- Add hold reference and rejection reason to bookings
ALTER TABLE bookings
    ADD COLUMN hold_id UUID,
    ADD COLUMN rejection_reason TEXT;

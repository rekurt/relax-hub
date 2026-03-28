CREATE TABLE IF NOT EXISTS booking_shares (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    token VARCHAR(32) NOT NULL UNIQUE,
    created_by UUID NOT NULL REFERENCES users(id),
    bathhouse_id UUID NOT NULL REFERENCES bathhouses(id),
    start_time TIMESTAMPTZ NOT NULL,
    end_time TIMESTAMPTZ NOT NULL,
    guest_count INT NOT NULL DEFAULT 1,
    booking_id UUID REFERENCES bookings(id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    expires_at TIMESTAMPTZ NOT NULL
);

CREATE INDEX idx_booking_shares_token ON booking_shares(token);
CREATE INDEX idx_booking_shares_expires_at ON booking_shares(expires_at);

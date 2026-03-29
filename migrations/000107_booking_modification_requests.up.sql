CREATE TABLE booking_modification_requests (
    id UUID PRIMARY KEY,
    booking_id UUID NOT NULL REFERENCES bookings(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    bathhouse_id UUID NOT NULL REFERENCES bathhouses(id) ON DELETE CASCADE,
    status TEXT NOT NULL DEFAULT 'pending',
    old_start_time TIMESTAMPTZ NOT NULL,
    old_end_time TIMESTAMPTZ NOT NULL,
    old_guest_count INTEGER NOT NULL,
    old_total_price BIGINT NOT NULL,
    proposed_start_time TIMESTAMPTZ NOT NULL,
    proposed_end_time TIMESTAMPTZ NOT NULL,
    proposed_guest_count INTEGER NOT NULL,
    proposed_total_price BIGINT NOT NULL,
    rejection_reason TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    expires_at TIMESTAMPTZ NOT NULL,
    resolved_at TIMESTAMPTZ
);

CREATE INDEX idx_booking_mod_requests_booking_id ON booking_modification_requests(booking_id);
CREATE INDEX idx_booking_mod_requests_status ON booking_modification_requests(status);
CREATE INDEX idx_booking_mod_requests_expires_at ON booking_modification_requests(expires_at) WHERE status = 'pending';

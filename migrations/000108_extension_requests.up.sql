CREATE TABLE IF NOT EXISTS extension_requests (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    booking_id UUID NOT NULL REFERENCES bookings(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    bathhouse_id UUID NOT NULL REFERENCES bathhouses(id) ON DELETE CASCADE,
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    extra_hours INTEGER NOT NULL,
    extension_price BIGINT NOT NULL,
    new_end_time TIMESTAMPTZ NOT NULL,
    hold_id UUID,
    rejection_reason TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    expires_at TIMESTAMPTZ NOT NULL,
    resolved_at TIMESTAMPTZ
);

CREATE INDEX idx_extension_requests_booking_id ON extension_requests(booking_id);
CREATE INDEX idx_extension_requests_status ON extension_requests(status);
CREATE INDEX idx_extension_requests_expires_at ON extension_requests(expires_at) WHERE status = 'pending';

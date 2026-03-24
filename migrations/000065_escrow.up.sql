CREATE TABLE escrows (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    booking_id UUID NOT NULL REFERENCES bookings(id) UNIQUE,
    amount BIGINT NOT NULL,
    service_fee BIGINT NOT NULL DEFAULT 0,
    status VARCHAR(20) NOT NULL DEFAULT 'held',
    claim_period_ends_at TIMESTAMPTZ NOT NULL,
    released_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_escrows_status ON escrows (status);
CREATE INDEX idx_escrows_claim_period ON escrows (claim_period_ends_at) WHERE status = 'held';

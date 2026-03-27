CREATE TABLE IF NOT EXISTS force_majeure_events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    admin_id UUID NOT NULL REFERENCES users(id),
    region VARCHAR(100) NOT NULL,
    date_from TIMESTAMPTZ NOT NULL,
    date_to TIMESTAMPTZ NOT NULL,
    reason TEXT NOT NULL,
    affected_count INT NOT NULL DEFAULT 0,
    total_refund BIGINT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_force_majeure_events_region ON force_majeure_events(region);
CREATE INDEX idx_force_majeure_events_created_at ON force_majeure_events(created_at DESC);

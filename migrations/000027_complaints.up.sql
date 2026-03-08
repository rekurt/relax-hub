CREATE TABLE IF NOT EXISTS complaints (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    reporter_id UUID NOT NULL REFERENCES users(id),
    target_type VARCHAR(20) NOT NULL CHECK (target_type IN ('review', 'bathhouse', 'user')),
    target_id UUID NOT NULL,
    reason VARCHAR(20) NOT NULL CHECK (reason IN ('spam', 'offensive', 'fake', 'fraud', 'other')),
    description TEXT NOT NULL DEFAULT '',
    status VARCHAR(20) NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'resolved', 'dismissed')),
    resolved_by_id UUID REFERENCES users(id),
    resolution TEXT NOT NULL DEFAULT '',
    resolved_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE(reporter_id, target_type, target_id)
);

CREATE INDEX idx_complaints_status ON complaints(status);
CREATE INDEX idx_complaints_target ON complaints(target_type, target_id);
CREATE INDEX idx_complaints_created_at ON complaints(created_at);

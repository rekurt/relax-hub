CREATE TABLE IF NOT EXISTS fraud_flags (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id),
    rule VARCHAR(50) NOT NULL,
    severity VARCHAR(20) NOT NULL DEFAULT 'low',
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    action VARCHAR(20) NOT NULL DEFAULT 'flag',
    details JSONB,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    reviewed_at TIMESTAMPTZ,
    reviewed_by UUID REFERENCES users(id)
);

CREATE INDEX idx_fraud_flags_user_id ON fraud_flags(user_id);
CREATE INDEX idx_fraud_flags_status ON fraud_flags(status);
CREATE INDEX idx_fraud_flags_rule ON fraud_flags(rule);
CREATE INDEX idx_fraud_flags_created_at ON fraud_flags(created_at DESC);
CREATE INDEX idx_fraud_flags_user_rule ON fraud_flags(user_id, rule);

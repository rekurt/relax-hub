-- Pricing Rules table
CREATE TABLE pricing_rules (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    bathhouse_id UUID NOT NULL REFERENCES bathhouses(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    type VARCHAR(30) NOT NULL,
    multiplier DECIMAL(5, 2) NOT NULL,
    days_of_week INTEGER[] DEFAULT NULL,
    time_from VARCHAR(5) DEFAULT NULL,
    time_to VARCHAR(5) DEFAULT NULL,
    date_from TIMESTAMPTZ DEFAULT NULL,
    date_to TIMESTAMPTZ DEFAULT NULL,
    priority INTEGER NOT NULL DEFAULT 0,
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_pricing_rules_bathhouse_id ON pricing_rules (bathhouse_id);
CREATE INDEX idx_pricing_rules_type ON pricing_rules (type);
CREATE INDEX idx_pricing_rules_active ON pricing_rules (is_active);
CREATE INDEX idx_pricing_rules_priority ON pricing_rules (priority);

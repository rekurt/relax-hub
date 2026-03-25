CREATE TABLE IF NOT EXISTS auto_scenarios (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    owner_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    type VARCHAR(50) NOT NULL,
    enabled BOOLEAN NOT NULL DEFAULT false,
    custom_text TEXT NOT NULL DEFAULT '',
    channel VARCHAR(20) NOT NULL DEFAULT 'push',
    delay_hours INTEGER NOT NULL DEFAULT 0,
    promo_code_id UUID REFERENCES promo_codes(id) ON DELETE SET NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(owner_id, type)
);

CREATE INDEX idx_auto_scenarios_owner_id ON auto_scenarios(owner_id);
CREATE INDEX idx_auto_scenarios_enabled ON auto_scenarios(enabled) WHERE enabled = true;

CREATE TABLE IF NOT EXISTS auto_scenario_executions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    scenario_id UUID NOT NULL REFERENCES auto_scenarios(id) ON DELETE CASCADE,
    guest_card_id UUID NOT NULL REFERENCES guest_cards(id) ON DELETE CASCADE,
    executed_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(scenario_id, guest_card_id)
);

CREATE INDEX idx_auto_scenario_executions_scenario_id ON auto_scenario_executions(scenario_id);

CREATE TABLE IF NOT EXISTS guest_cards (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    owner_id UUID NOT NULL REFERENCES users(id),
    client_id UUID NOT NULL REFERENCES users(id),
    bathhouse_id UUID NOT NULL REFERENCES bathhouses(id),
    first_visit_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    last_visit_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    visit_count INT NOT NULL DEFAULT 1,
    total_spent BIGINT NOT NULL DEFAULT 0,
    avg_check BIGINT NOT NULL DEFAULT 0,
    notes TEXT NOT NULL DEFAULT '',
    tags TEXT[] NOT NULL DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (owner_id, client_id, bathhouse_id)
);

CREATE INDEX idx_guest_cards_owner_id ON guest_cards(owner_id);
CREATE INDEX idx_guest_cards_client_id ON guest_cards(client_id);
CREATE INDEX idx_guest_cards_bathhouse_id ON guest_cards(bathhouse_id);
CREATE INDEX idx_guest_cards_last_visit_at ON guest_cards(last_visit_at);
CREATE INDEX idx_guest_cards_total_spent ON guest_cards(total_spent);

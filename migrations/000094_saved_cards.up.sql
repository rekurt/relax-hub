CREATE TABLE saved_cards (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    provider_token VARCHAR(500) NOT NULL,
    last4 VARCHAR(4) NOT NULL,
    brand VARCHAR(50) NOT NULL,
    expiry_month INTEGER NOT NULL CHECK (expiry_month >= 1 AND expiry_month <= 12),
    expiry_year INTEGER NOT NULL CHECK (expiry_year >= 2024),
    is_default BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_saved_cards_user ON saved_cards(user_id);
CREATE UNIQUE INDEX idx_saved_cards_user_default ON saved_cards(user_id) WHERE is_default = true;

CREATE TABLE IF NOT EXISTS broadcasts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    owner_id UUID NOT NULL REFERENCES users(id),
    segment VARCHAR(50) NOT NULL,
    title VARCHAR(255) NOT NULL,
    body TEXT NOT NULL,
    image_url TEXT NOT NULL DEFAULT '',
    promo_code_id UUID REFERENCES promo_codes(id),
    channels TEXT[] NOT NULL DEFAULT '{}',
    status VARCHAR(20) NOT NULL DEFAULT 'draft',
    delivered BIGINT NOT NULL DEFAULT 0,
    read BIGINT NOT NULL DEFAULT 0,
    sent_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_broadcasts_owner_id ON broadcasts(owner_id);
CREATE INDEX idx_broadcasts_owner_status ON broadcasts(owner_id, status);
CREATE INDEX idx_broadcasts_owner_created ON broadcasts(owner_id, created_at DESC);

-- Client reviews: owner rates client (bidirectional reviews, FR-130/FR-131)
CREATE TABLE client_reviews (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    owner_id UUID NOT NULL REFERENCES users(id),
    client_id UUID NOT NULL REFERENCES users(id),
    booking_id UUID NOT NULL REFERENCES bookings(id),
    bathhouse_id UUID NOT NULL REFERENCES bathhouses(id),
    punctuality DECIMAL(2,1) CHECK (punctuality >= 1.0 AND punctuality <= 5.0 AND punctuality * 2 = FLOOR(punctuality * 2)),
    cleanliness DECIMAL(2,1) CHECK (cleanliness >= 1.0 AND cleanliness <= 5.0 AND cleanliness * 2 = FLOOR(cleanliness * 2)),
    rule_compliance DECIMAL(2,1) CHECK (rule_compliance >= 1.0 AND rule_compliance <= 5.0 AND rule_compliance * 2 = FLOOR(rule_compliance * 2)),
    rating INT NOT NULL CHECK (rating >= 1 AND rating <= 5),
    text TEXT NOT NULL DEFAULT '',
    reveal_at TIMESTAMPTZ NOT NULL,
    is_revealed BOOLEAN NOT NULL DEFAULT false,
    moderation_score DECIMAL(3,2),
    moderation_flags TEXT[] DEFAULT '{}',
    status VARCHAR(20) NOT NULL DEFAULT 'approved',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT uq_client_reviews_booking UNIQUE (booking_id)
);

CREATE INDEX idx_client_reviews_client_id ON client_reviews (client_id);
CREATE INDEX idx_client_reviews_owner_id ON client_reviews (owner_id);
CREATE INDEX idx_client_reviews_bathhouse_id ON client_reviews (bathhouse_id);
CREATE INDEX idx_client_reviews_reveal_at ON client_reviews (reveal_at) WHERE is_revealed = false;

-- Add reveal_at to existing reviews table for double-blind support
ALTER TABLE reviews ADD COLUMN IF NOT EXISTS reveal_at TIMESTAMPTZ;
ALTER TABLE reviews ADD COLUMN IF NOT EXISTS is_revealed BOOLEAN NOT NULL DEFAULT true;

-- Existing reviews are already revealed
UPDATE reviews SET is_revealed = true WHERE is_revealed = true;

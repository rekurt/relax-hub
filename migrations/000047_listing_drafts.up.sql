CREATE TABLE IF NOT EXISTS listing_drafts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    status VARCHAR(20) NOT NULL DEFAULT 'draft',
    current_step INTEGER NOT NULL DEFAULT 1,
    step_data JSONB NOT NULL DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_listing_drafts_user_id ON listing_drafts(user_id);
CREATE INDEX idx_listing_drafts_status ON listing_drafts(status);

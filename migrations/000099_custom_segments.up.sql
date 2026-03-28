CREATE TABLE custom_segments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    owner_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    bathhouse_id UUID REFERENCES bathhouses(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    conditions JSONB NOT NULL DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_custom_segments_owner ON custom_segments(owner_id);
CREATE INDEX idx_custom_segments_bathhouse ON custom_segments(bathhouse_id) WHERE bathhouse_id IS NOT NULL;

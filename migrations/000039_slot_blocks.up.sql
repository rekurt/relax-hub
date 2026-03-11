CREATE TABLE IF NOT EXISTS slot_blocks (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    bathhouse_id UUID NOT NULL REFERENCES bathhouses(id) ON DELETE CASCADE,
    start_time TIMESTAMPTZ NOT NULL,
    end_time TIMESTAMPTZ NOT NULL,
    source VARCHAR(50) NOT NULL DEFAULT 'manual',
    external_id VARCHAR(255) NOT NULL DEFAULT '',
    description TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT slot_blocks_end_after_start CHECK (end_time > start_time)
);

CREATE INDEX idx_slot_blocks_bathhouse_id ON slot_blocks(bathhouse_id);
CREATE INDEX idx_slot_blocks_bathhouse_time ON slot_blocks(bathhouse_id, start_time, end_time);
CREATE INDEX idx_slot_blocks_external ON slot_blocks(bathhouse_id, source, external_id) WHERE external_id != '';

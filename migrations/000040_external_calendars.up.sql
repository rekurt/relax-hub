CREATE TABLE external_calendars (
    id UUID PRIMARY KEY,
    bathhouse_id UUID NOT NULL REFERENCES bathhouses(id) ON DELETE CASCADE,
    url TEXT NOT NULL,
    source VARCHAR(50) NOT NULL,
    last_sync_at TIMESTAMPTZ,
    last_error TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_external_calendars_bathhouse_id ON external_calendars(bathhouse_id);

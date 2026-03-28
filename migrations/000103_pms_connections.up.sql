CREATE TABLE pms_connections (
    id UUID PRIMARY KEY,
    owner_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    bathhouse_id UUID NOT NULL REFERENCES bathhouses(id) ON DELETE CASCADE,
    provider TEXT NOT NULL,
    credentials_encrypted TEXT NOT NULL,
    sync_direction TEXT NOT NULL DEFAULT 'both',
    sync_interval_minutes INTEGER NOT NULL DEFAULT 15,
    status TEXT NOT NULL DEFAULT 'active',
    last_sync_at TIMESTAMP,
    last_sync_error TEXT NOT NULL DEFAULT '',
    external_id TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX idx_pms_connections_bathhouse_id ON pms_connections(bathhouse_id);
CREATE INDEX idx_pms_connections_owner_id ON pms_connections(owner_id);
CREATE INDEX idx_pms_connections_status ON pms_connections(status) WHERE status = 'active';

CREATE TABLE pms_sync_logs (
    id UUID PRIMARY KEY,
    connection_id UUID NOT NULL REFERENCES pms_connections(id) ON DELETE CASCADE,
    direction TEXT NOT NULL,
    status TEXT NOT NULL,
    items_synced INTEGER NOT NULL DEFAULT 0,
    error_message TEXT NOT NULL DEFAULT '',
    started_at TIMESTAMP NOT NULL DEFAULT NOW(),
    completed_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_pms_sync_logs_connection_id ON pms_sync_logs(connection_id);
CREATE INDEX idx_pms_sync_logs_started_at ON pms_sync_logs(started_at);

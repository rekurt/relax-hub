CREATE TABLE photo_orders (
    id UUID PRIMARY KEY,
    owner_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    bathhouse_id UUID NOT NULL REFERENCES bathhouses(id) ON DELETE CASCADE,
    region TEXT NOT NULL DEFAULT 'RU',
    status TEXT NOT NULL DEFAULT 'requested',
    photographer_name TEXT NOT NULL DEFAULT '',
    price BIGINT NOT NULL DEFAULT 0,
    scheduled_at TIMESTAMP,
    notes TEXT NOT NULL DEFAULT '',
    admin_notes TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_photo_orders_owner_id ON photo_orders(owner_id);
CREATE INDEX idx_photo_orders_bathhouse_id ON photo_orders(bathhouse_id);
CREATE INDEX idx_photo_orders_status ON photo_orders(status);
CREATE INDEX idx_photo_orders_created_at ON photo_orders(created_at DESC);

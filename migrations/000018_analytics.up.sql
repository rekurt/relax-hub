-- Bathhouse views tracking
CREATE TABLE bathhouse_views (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    bathhouse_id UUID NOT NULL REFERENCES bathhouses(id) ON DELETE CASCADE,
    viewer_id UUID REFERENCES users(id) ON DELETE SET NULL,
    source VARCHAR(20) NOT NULL DEFAULT 'direct',
    ip_hash VARCHAR(64) NOT NULL,
    viewed_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_bathhouse_views_bathhouse_id ON bathhouse_views (bathhouse_id, viewed_at DESC);
CREATE INDEX idx_bathhouse_views_viewer_id ON bathhouse_views (viewer_id) WHERE viewer_id IS NOT NULL;
CREATE INDEX idx_bathhouse_views_ip_hash ON bathhouse_views (bathhouse_id, ip_hash, viewed_at DESC);
CREATE INDEX idx_bathhouse_views_viewed_at ON bathhouse_views (viewed_at);

-- Analytics snapshots (daily aggregated data)
CREATE TABLE analytics_snapshots (
    bathhouse_id UUID NOT NULL REFERENCES bathhouses(id) ON DELETE CASCADE,
    date DATE NOT NULL,
    views BIGINT NOT NULL DEFAULT 0,
    unique_views BIGINT NOT NULL DEFAULT 0,
    bookings BIGINT NOT NULL DEFAULT 0,
    revenue BIGINT NOT NULL DEFAULT 0,
    review_count INT NOT NULL DEFAULT 0,
    avg_rating DOUBLE PRECISION NOT NULL DEFAULT 0,
    PRIMARY KEY (bathhouse_id, date)
);

CREATE INDEX idx_analytics_snapshots_date ON analytics_snapshots (date);

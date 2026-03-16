CREATE TABLE IF NOT EXISTS bathhouse_photos (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    bathhouse_id UUID NOT NULL REFERENCES bathhouses(id) ON DELETE CASCADE,
    url TEXT NOT NULL,
    thumbnail_url TEXT NOT NULL DEFAULT '',
    position INT NOT NULL DEFAULT 0,
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    verified_by_id UUID REFERENCES users(id) ON DELETE SET NULL,
    verified_at TIMESTAMPTZ,
    rejection_reason TEXT NOT NULL DEFAULT '',
    uploaded_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_bathhouse_photos_bathhouse_id ON bathhouse_photos(bathhouse_id);
CREATE INDEX idx_bathhouse_photos_status ON bathhouse_photos(status);

ALTER TABLE bathhouses ADD COLUMN IF NOT EXISTS is_photo_verified BOOLEAN NOT NULL DEFAULT false;

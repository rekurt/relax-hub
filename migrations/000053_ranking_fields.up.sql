-- Add ranking fields to bathhouses for composite search ranking
ALTER TABLE bathhouses ADD COLUMN IF NOT EXISTS conversion_rate DOUBLE PRECISION NOT NULL DEFAULT 0;
ALTER TABLE bathhouses ADD COLUMN IF NOT EXISTS occupancy_rate DOUBLE PRECISION NOT NULL DEFAULT 0;
ALTER TABLE bathhouses ADD COLUMN IF NOT EXISTS view_count BIGINT NOT NULL DEFAULT 0;

-- Index for sorting by view_count (popularity)
CREATE INDEX IF NOT EXISTS idx_bathhouses_view_count ON bathhouses (view_count DESC);

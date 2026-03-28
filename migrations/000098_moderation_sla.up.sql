-- Add moderation tracking fields to reviews table for SLA metrics
ALTER TABLE reviews ADD COLUMN IF NOT EXISTS moderated_by UUID REFERENCES users(id) ON DELETE SET NULL;
ALTER TABLE reviews ADD COLUMN IF NOT EXISTS moderated_at TIMESTAMPTZ;

CREATE INDEX idx_reviews_moderated_by ON reviews (moderated_by) WHERE moderated_by IS NOT NULL;
CREATE INDEX idx_reviews_status_created ON reviews (status, created_at) WHERE status = 'pending';

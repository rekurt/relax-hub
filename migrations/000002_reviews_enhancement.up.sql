-- Add new columns to reviews table for enhanced review functionality
ALTER TABLE reviews ADD COLUMN status VARCHAR(20) NOT NULL DEFAULT 'approved';
ALTER TABLE reviews ADD COLUMN owner_response TEXT NOT NULL DEFAULT '';
ALTER TABLE reviews ADD COLUMN owner_response_at TIMESTAMPTZ;
ALTER TABLE reviews ADD COLUMN updated_at TIMESTAMPTZ NOT NULL DEFAULT now();
ALTER TABLE reviews ADD COLUMN images JSONB NOT NULL DEFAULT '[]';

CREATE INDEX idx_reviews_status ON reviews (status);

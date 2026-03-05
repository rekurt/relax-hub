-- Add rejection_reasons column to reviews table for moderation
ALTER TABLE reviews ADD COLUMN IF NOT EXISTS rejection_reasons TEXT[] NOT NULL DEFAULT '{}'::TEXT[];

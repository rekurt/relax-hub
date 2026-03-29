-- Add daily_bid_kopecks column to promotions table for auction-based promotion ranking
ALTER TABLE promotions ADD COLUMN IF NOT EXISTS daily_bid_kopecks BIGINT NOT NULL DEFAULT 5000;

-- Add index for fast lookup of active promotions with bid amount
CREATE INDEX IF NOT EXISTS idx_promotions_active_bid ON promotions (status, daily_bid_kopecks DESC) WHERE status = 'active';

DROP INDEX IF EXISTS idx_promotions_active_bid;
ALTER TABLE promotions DROP COLUMN IF EXISTS daily_bid_kopecks;

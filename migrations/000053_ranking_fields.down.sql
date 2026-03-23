DROP INDEX IF EXISTS idx_bathhouses_view_count;
ALTER TABLE bathhouses DROP COLUMN IF EXISTS view_count;
ALTER TABLE bathhouses DROP COLUMN IF EXISTS occupancy_rate;
ALTER TABLE bathhouses DROP COLUMN IF EXISTS conversion_rate;

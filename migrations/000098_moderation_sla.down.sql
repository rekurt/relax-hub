DROP INDEX IF EXISTS idx_reviews_moderated_by;
DROP INDEX IF EXISTS idx_reviews_status_created;
ALTER TABLE reviews DROP COLUMN IF EXISTS moderated_at;
ALTER TABLE reviews DROP COLUMN IF EXISTS moderated_by;

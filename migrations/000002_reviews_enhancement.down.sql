DROP INDEX IF EXISTS idx_reviews_status;

ALTER TABLE reviews DROP COLUMN IF EXISTS images;
ALTER TABLE reviews DROP COLUMN IF EXISTS updated_at;
ALTER TABLE reviews DROP COLUMN IF EXISTS owner_response_at;
ALTER TABLE reviews DROP COLUMN IF EXISTS owner_response;
ALTER TABLE reviews DROP COLUMN IF EXISTS status;

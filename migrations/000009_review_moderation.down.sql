-- Remove rejection_reasons column
ALTER TABLE reviews DROP COLUMN IF EXISTS rejection_reasons;

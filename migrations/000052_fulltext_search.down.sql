DROP TRIGGER IF EXISTS bathhouses_search_vector_trigger ON bathhouses;
DROP FUNCTION IF EXISTS bathhouses_search_vector_update();
DROP INDEX IF EXISTS idx_bathhouses_description_trgm;
DROP INDEX IF EXISTS idx_bathhouses_name_trgm;
DROP INDEX IF EXISTS idx_bathhouses_search_vector;
ALTER TABLE bathhouses DROP COLUMN IF EXISTS search_vector;

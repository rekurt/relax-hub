-- Add slug column to bathhouses
ALTER TABLE bathhouses ADD COLUMN slug VARCHAR(255);

-- Generate slugs for existing bathhouses from their names
-- Uses a simple transliteration approach: lowercase, replace spaces with dashes
-- This is a basic version; the application handles proper transliteration
UPDATE bathhouses SET slug = LOWER(REPLACE(REPLACE(TRIM(name), ' ', '-'), '.', '')) WHERE slug IS NULL;

-- Ensure uniqueness by appending ID suffix for any duplicates
WITH duplicates AS (
    SELECT id, slug, ROW_NUMBER() OVER (PARTITION BY slug ORDER BY created_at) as rn
    FROM bathhouses
)
UPDATE bathhouses SET slug = bathhouses.slug || '-' || SUBSTRING(bathhouses.id::text, 1, 8)
FROM duplicates
WHERE bathhouses.id = duplicates.id AND duplicates.rn > 1;

-- Now make slug NOT NULL and UNIQUE
ALTER TABLE bathhouses ALTER COLUMN slug SET NOT NULL;
CREATE UNIQUE INDEX idx_bathhouses_slug ON bathhouses (slug);

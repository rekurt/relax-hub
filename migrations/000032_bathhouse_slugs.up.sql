-- Add slug column to bathhouses
ALTER TABLE bathhouses ADD COLUMN slug VARCHAR(255);

-- Generate slugs for existing bathhouses from their names
-- Transliterate Russian characters to Latin, lowercase, replace non-alphanumeric with dashes
UPDATE bathhouses SET slug =
    TRIM(BOTH '-' FROM
        REGEXP_REPLACE(
            REGEXP_REPLACE(
                LOWER(
                    REPLACE(REPLACE(REPLACE(REPLACE(REPLACE(REPLACE(REPLACE(REPLACE(REPLACE(REPLACE(
                    REPLACE(REPLACE(REPLACE(REPLACE(REPLACE(REPLACE(REPLACE(REPLACE(REPLACE(REPLACE(
                    REPLACE(REPLACE(REPLACE(REPLACE(REPLACE(REPLACE(REPLACE(REPLACE(REPLACE(REPLACE(
                    REPLACE(REPLACE(REPLACE(
                    REPLACE(REPLACE(REPLACE(REPLACE(REPLACE(REPLACE(REPLACE(REPLACE(REPLACE(REPLACE(
                    REPLACE(REPLACE(REPLACE(REPLACE(REPLACE(REPLACE(REPLACE(REPLACE(REPLACE(REPLACE(
                    REPLACE(REPLACE(REPLACE(REPLACE(REPLACE(REPLACE(REPLACE(REPLACE(REPLACE(REPLACE(
                    REPLACE(REPLACE(REPLACE(
                    TRIM(name),
                    'а', 'a'), 'б', 'b'), 'в', 'v'), 'г', 'g'), 'д', 'd'),
                    'е', 'e'), 'ё', 'yo'), 'ж', 'zh'), 'з', 'z'), 'и', 'i'),
                    'й', 'y'), 'к', 'k'), 'л', 'l'), 'м', 'm'), 'н', 'n'),
                    'о', 'o'), 'п', 'p'), 'р', 'r'), 'с', 's'), 'т', 't'),
                    'у', 'u'), 'ф', 'f'), 'х', 'kh'), 'ц', 'ts'), 'ч', 'ch'),
                    'ш', 'sh'), 'щ', 'shch'), 'ъ', ''), 'ы', 'y'), 'ь', ''),
                    'э', 'e'), 'ю', 'yu'), 'я', 'ya'),
                    'А', 'a'), 'Б', 'b'), 'В', 'v'), 'Г', 'g'), 'Д', 'd'),
                    'Е', 'e'), 'Ё', 'yo'), 'Ж', 'zh'), 'З', 'z'), 'И', 'i'),
                    'Й', 'y'), 'К', 'k'), 'Л', 'l'), 'М', 'm'), 'Н', 'n'),
                    'О', 'o'), 'П', 'p'), 'Р', 'r'), 'С', 's'), 'Т', 't'),
                    'У', 'u'), 'Ф', 'f'), 'Х', 'kh'), 'Ц', 'ts'), 'Ч', 'ch'),
                    'Ш', 'sh'), 'Щ', 'shch'), 'Ъ', ''), 'Ы', 'y'), 'Ь', ''),
                    'Э', 'e'), 'Ю', 'yu'), 'Я', 'ya')
                ),
            '[^a-z0-9]+', '-', 'g'),
        '-{2,}', '-', 'g')
    )
WHERE slug IS NULL;

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

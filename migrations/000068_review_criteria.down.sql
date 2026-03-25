ALTER TABLE reviews
    DROP CONSTRAINT IF EXISTS chk_cleanliness,
    DROP CONSTRAINT IF EXISTS chk_accuracy,
    DROP CONSTRAINT IF EXISTS chk_communication,
    DROP CONSTRAINT IF EXISTS chk_value_for_money;

ALTER TABLE reviews
    DROP COLUMN IF EXISTS cleanliness,
    DROP COLUMN IF EXISTS accuracy,
    DROP COLUMN IF EXISTS communication,
    DROP COLUMN IF EXISTS value_for_money;

-- Fix UNIQUE (region, category) not deduplicating NULL category values in PostgreSQL.
-- Drop the table-level constraint and replace with two indexes:
-- 1. A partial unique index for rows where category IS NULL
-- 2. A regular unique index for rows where category IS NOT NULL
ALTER TABLE service_fee_configs DROP CONSTRAINT IF EXISTS service_fee_configs_region_category_key;

CREATE UNIQUE INDEX IF NOT EXISTS idx_service_fee_region_no_category
    ON service_fee_configs (region) WHERE category IS NULL;

CREATE UNIQUE INDEX IF NOT EXISTS idx_service_fee_region_category
    ON service_fee_configs (region, category) WHERE category IS NOT NULL;

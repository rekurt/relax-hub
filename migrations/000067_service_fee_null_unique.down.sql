DROP INDEX IF EXISTS idx_service_fee_region_no_category;
DROP INDEX IF EXISTS idx_service_fee_region_category;

ALTER TABLE service_fee_configs ADD CONSTRAINT service_fee_configs_region_category_key UNIQUE (region, category);

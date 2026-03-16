CREATE EXTENSION IF NOT EXISTS pgcrypto;

ALTER TABLE bathhouses ADD COLUMN api_key VARCHAR(64) UNIQUE;

-- Populate existing bathhouses with generated API keys
UPDATE bathhouses SET api_key = encode(gen_random_bytes(32), 'hex') WHERE api_key IS NULL;

-- Add NOT NULL constraint
ALTER TABLE bathhouses ALTER COLUMN api_key SET NOT NULL;

CREATE INDEX idx_bathhouses_api_key ON bathhouses(api_key);

ALTER TABLE bathhouses ADD COLUMN api_key VARCHAR(64) UNIQUE;

CREATE INDEX idx_bathhouses_api_key ON bathhouses(api_key);

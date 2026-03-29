DROP INDEX IF EXISTS idx_payouts_external_id;
ALTER TABLE payouts DROP COLUMN IF EXISTS external_id;
ALTER TABLE payouts DROP COLUMN IF EXISTS payout_method;

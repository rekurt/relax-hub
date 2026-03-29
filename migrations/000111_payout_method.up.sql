ALTER TABLE payouts ADD COLUMN payout_method VARCHAR(20) NOT NULL DEFAULT 'bank_transfer';
ALTER TABLE payouts ADD COLUMN external_id VARCHAR(255);

CREATE INDEX idx_payouts_external_id ON payouts(external_id) WHERE external_id IS NOT NULL;

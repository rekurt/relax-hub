-- Add 'refund' to loyalty_transactions type check constraint
ALTER TABLE loyalty_transactions DROP CONSTRAINT IF EXISTS loyalty_transactions_type_check;
ALTER TABLE loyalty_transactions ADD CONSTRAINT loyalty_transactions_type_check CHECK (type IN ('earn', 'spend', 'refund'));

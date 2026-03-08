ALTER TABLE referral_balances ADD CONSTRAINT referral_balances_balance_non_negative CHECK (balance >= 0);

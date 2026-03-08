DROP TABLE IF EXISTS referral_balances;
DROP TABLE IF EXISTS referrals;
ALTER TABLE users DROP COLUMN IF EXISTS referral_code;

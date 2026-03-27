CREATE TABLE IF NOT EXISTS platform_settings (
    key         VARCHAR(100) PRIMARY KEY,
    value       TEXT         NOT NULL DEFAULT '',
    description TEXT         NOT NULL DEFAULT '',
    type        VARCHAR(20)  NOT NULL DEFAULT 'string',
    updated_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_by  UUID
);

-- Seed initial settings
INSERT INTO platform_settings (key, value, description, type) VALUES
    ('service_fee_percent', '10', 'Default platform service fee percentage', 'float'),
    ('welcome_bonus_amount', '50000', 'Welcome bonus amount in kopecks', 'int'),
    ('welcome_bonus_expiry_days', '30', 'Welcome bonus expiration in days', 'int'),
    ('wallet_bonus_expiry_days', '180', 'Wallet bonus expiration in days', 'int'),
    ('wallet_refund_bonus_percent', '5', 'Bonus percentage for wallet refund', 'int'),
    ('max_wallet_balance', '10000000', 'Maximum wallet balance in kopecks', 'int'),
    ('escrow_claim_hours', '48', 'Escrow hold period before release to owner (hours)', 'int'),
    ('min_payout_amount', '100000', 'Minimum payout amount in kopecks', 'int'),
    ('bayesian_min_reviews', '5', 'Minimum reviews for Bayesian average', 'int'),
    ('review_request_delay_hours', '2', 'Delay before sending review request after checkout (hours)', 'int'),
    ('noshow_grace_minutes', '30', 'Grace period for no-show detection (minutes)', 'int'),
    ('offer_version', '1.0', 'Current offer/contract version', 'string')
ON CONFLICT (key) DO NOTHING;

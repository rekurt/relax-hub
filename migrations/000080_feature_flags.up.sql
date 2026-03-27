CREATE TABLE IF NOT EXISTS feature_flags (
    key         VARCHAR(100) PRIMARY KEY,
    enabled     BOOLEAN NOT NULL DEFAULT FALSE,
    description VARCHAR(500) NOT NULL DEFAULT '',
    region      VARCHAR(100),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_by  UUID REFERENCES users(id)
);

INSERT INTO feature_flags (key, description) VALUES
    ('wallet_enabled', 'Enable wallet system'),
    ('phone_auth_enabled', 'Enable phone + OTP authentication'),
    ('two_fa_enabled', 'Enable two-factor authentication'),
    ('kyc_required', 'Require KYC verification for owners'),
    ('addons_enabled', 'Enable booking add-ons'),
    ('request_booking_enabled', 'Enable request-based booking mode'),
    ('escrow_enabled', 'Enable escrow payment holds'),
    ('crm_enabled', 'Enable CRM features for owners'),
    ('disputes_enabled', 'Enable dispute system'),
    ('antifraud_enabled', 'Enable anti-fraud engine'),
    ('sbp_payments_enabled', 'Enable SBP payment method'),
    ('last_minute_enabled', 'Enable last-minute discount pricing'),
    ('fulltext_search_enabled', 'Enable full-text search with tsvector')
ON CONFLICT (key) DO NOTHING;

CREATE TABLE wallets (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id),
    balance BIGINT NOT NULL DEFAULT 0,
    held_amount BIGINT NOT NULL DEFAULT 0,
    currency VARCHAR(3) NOT NULL DEFAULT 'RUB',
    status VARCHAR(20) NOT NULL DEFAULT 'active',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE(user_id)
);

CREATE INDEX idx_wallets_user_id ON wallets(user_id);

CREATE TABLE wallet_transactions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    wallet_id UUID NOT NULL REFERENCES wallets(id),
    type VARCHAR(30) NOT NULL,
    amount BIGINT NOT NULL,
    balance_after BIGINT NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'completed',
    description TEXT,
    reference_type VARCHAR(30),
    reference_id UUID,
    is_bonus BOOLEAN NOT NULL DEFAULT false,
    expires_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_wallet_tx_wallet_type ON wallet_transactions(wallet_id, type, created_at);
CREATE INDEX idx_wallet_tx_expires ON wallet_transactions(wallet_id, expires_at)
    WHERE expires_at IS NOT NULL AND is_bonus = true;
CREATE INDEX idx_wallet_tx_wallet_created ON wallet_transactions(wallet_id, created_at DESC);

CREATE TABLE wallet_holds (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    wallet_id UUID NOT NULL REFERENCES wallets(id),
    amount BIGINT NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'active',
    description TEXT,
    reference_type VARCHAR(30),
    reference_id UUID,
    expires_at TIMESTAMPTZ NOT NULL,
    captured_at TIMESTAMPTZ,
    released_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_wallet_holds_wallet_status ON wallet_holds(wallet_id, status);
CREATE INDEX idx_wallet_holds_expires ON wallet_holds(expires_at) WHERE status = 'active';

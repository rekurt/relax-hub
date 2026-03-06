-- Loyalty accounts table
CREATE TABLE loyalty_accounts (
    user_id UUID PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    level VARCHAR(30) NOT NULL DEFAULT 'bronze',
    points BIGINT NOT NULL DEFAULT 0 CHECK (points >= 0),
    total_earned BIGINT NOT NULL DEFAULT 0 CHECK (total_earned >= 0),
    total_spent BIGINT NOT NULL DEFAULT 0 CHECK (total_spent >= 0),
    visit_count INTEGER NOT NULL DEFAULT 0 CHECK (visit_count >= 0),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_loyalty_accounts_level ON loyalty_accounts (level);

-- Loyalty transactions table
CREATE TABLE loyalty_transactions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    type VARCHAR(10) NOT NULL CHECK (type IN ('earn', 'spend')),
    amount BIGINT NOT NULL CHECK (amount > 0),
    booking_id UUID REFERENCES bookings(id) ON DELETE SET NULL,
    description TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_loyalty_transactions_user_id ON loyalty_transactions (user_id);
CREATE INDEX idx_loyalty_transactions_booking_id ON loyalty_transactions (booking_id);
CREATE INDEX idx_loyalty_transactions_created_at ON loyalty_transactions (created_at);

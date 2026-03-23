CREATE TABLE owner_payment_details (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    entity_type VARCHAR(30) NOT NULL,
    bank_card_number VARCHAR(16),
    card_holder_name VARCHAR(255),
    bank_account VARCHAR(20),
    bik VARCHAR(9),
    inn VARCHAR(12),
    correspondent_account VARCHAR(20),
    bank_name VARCHAR(255),
    is_verified BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT owner_payment_details_user_unique UNIQUE (user_id)
);

CREATE INDEX idx_owner_payment_details_user_id ON owner_payment_details(user_id);

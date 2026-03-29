CREATE TABLE antifraud_stoplist (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    phone VARCHAR(20),
    email VARCHAR(255),
    inn VARCHAR(12),
    bank_card_number VARCHAR(16),
    reason TEXT NOT NULL,
    blocked_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_by UUID REFERENCES users(id),
    CONSTRAINT antifraud_stoplist_has_identifier CHECK (
        phone IS NOT NULL OR email IS NOT NULL OR inn IS NOT NULL OR bank_card_number IS NOT NULL
    )
);

CREATE INDEX idx_antifraud_stoplist_phone ON antifraud_stoplist (phone) WHERE phone IS NOT NULL;
CREATE INDEX idx_antifraud_stoplist_email ON antifraud_stoplist (email) WHERE email IS NOT NULL;
CREATE INDEX idx_antifraud_stoplist_inn ON antifraud_stoplist (inn) WHERE inn IS NOT NULL;
CREATE INDEX idx_antifraud_stoplist_bank_card ON antifraud_stoplist (bank_card_number) WHERE bank_card_number IS NOT NULL;

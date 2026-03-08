CREATE TABLE IF NOT EXISTS gift_certificates (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    code VARCHAR(14) NOT NULL UNIQUE,
    purchaser_id UUID REFERENCES users(id),
    purchaser_email VARCHAR(255) NOT NULL,
    recipient_email VARCHAR(255) NOT NULL,
    recipient_name VARCHAR(255) NOT NULL DEFAULT '',
    amount BIGINT NOT NULL CHECK (amount > 0),
    balance BIGINT NOT NULL CHECK (balance >= 0),
    message TEXT NOT NULL DEFAULT '',
    status VARCHAR(20) NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'used', 'expired')),
    valid_until TIMESTAMPTZ NOT NULL,
    redeemed_by_id UUID REFERENCES users(id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_gift_certificates_purchaser_id ON gift_certificates(purchaser_id);
CREATE INDEX idx_gift_certificates_redeemed_by_id ON gift_certificates(redeemed_by_id);
CREATE INDEX idx_gift_certificates_status ON gift_certificates(status);

CREATE TABLE IF NOT EXISTS certificate_usages (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    certificate_id UUID NOT NULL REFERENCES gift_certificates(id),
    booking_id UUID NOT NULL REFERENCES bookings(id),
    amount BIGINT NOT NULL CHECK (amount > 0),
    used_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_certificate_usages_certificate_id ON certificate_usages(certificate_id);
CREATE INDEX idx_certificate_usages_booking_id ON certificate_usages(booking_id);

-- PromoCode table
CREATE TABLE promo_codes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    code VARCHAR(50) NOT NULL UNIQUE,
    type VARCHAR(30) NOT NULL,
    value BIGINT NOT NULL,
    bathhouse_id UUID REFERENCES bathhouses(id) ON DELETE CASCADE,
    creator_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    max_uses INTEGER NOT NULL DEFAULT 0,
    current_uses INTEGER NOT NULL DEFAULT 0,
    min_amount BIGINT NOT NULL DEFAULT 0,
    valid_from TIMESTAMPTZ NOT NULL,
    valid_until TIMESTAMPTZ NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_promo_codes_bathhouse_id ON promo_codes (bathhouse_id);
CREATE INDEX idx_promo_codes_creator_id ON promo_codes (creator_id);
CREATE INDEX idx_promo_codes_is_active ON promo_codes (is_active);
CREATE INDEX idx_promo_codes_valid_until ON promo_codes (valid_until);

-- PromoUsage table
CREATE TABLE promo_usages (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    promo_code_id UUID NOT NULL REFERENCES promo_codes(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    booking_id UUID NOT NULL REFERENCES bookings(id) ON DELETE CASCADE,
    discount_amount BIGINT NOT NULL,
    used_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_promo_usages_promo_code_id ON promo_usages (promo_code_id);
CREATE INDEX idx_promo_usages_user_id ON promo_usages (user_id);
CREATE INDEX idx_promo_usages_booking_id ON promo_usages (booking_id);
CREATE INDEX idx_promo_usages_used_at ON promo_usages (used_at);

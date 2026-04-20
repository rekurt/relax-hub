CREATE TABLE IF NOT EXISTS certificate_orders (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    purchaser_id UUID REFERENCES users(id),
    purchaser_email TEXT NOT NULL,
    recipient_email TEXT NOT NULL DEFAULT '',
    recipient_name TEXT NOT NULL DEFAULT '',
    message TEXT NOT NULL DEFAULT '',
    amount BIGINT NOT NULL,
    status VARCHAR(32) NOT NULL DEFAULT 'draft',
    payment_method VARCHAR(32),
    provider VARCHAR(32),
    external_id VARCHAR(255),
    certificate_id UUID REFERENCES gift_certificates(id),
    paid_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_certificate_orders_external_id
    ON certificate_orders(external_id)
    WHERE external_id IS NOT NULL;

CREATE INDEX IF NOT EXISTS idx_certificate_orders_purchaser_id
    ON certificate_orders(purchaser_id);

CREATE INDEX IF NOT EXISTS idx_certificate_orders_status
    ON certificate_orders(status);

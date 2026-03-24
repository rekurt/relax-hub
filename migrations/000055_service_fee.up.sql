CREATE TABLE service_fee_configs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    region VARCHAR(10) NOT NULL DEFAULT '*',
    category VARCHAR(100),
    fee_percent NUMERIC(5,2) NOT NULL DEFAULT 10.0 CHECK (fee_percent >= 0 AND fee_percent <= 25),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (region, category)
);

INSERT INTO service_fee_configs (region, fee_percent) VALUES ('*', 10.0);

ALTER TABLE bookings ADD COLUMN service_fee_amount BIGINT NOT NULL DEFAULT 0;

-- Subscriptions table
CREATE TABLE subscriptions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    bathhouse_id UUID NOT NULL REFERENCES bathhouses(id) ON DELETE CASCADE,
    owner_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    plan VARCHAR(30) NOT NULL,
    status VARCHAR(30) NOT NULL,
    start_date TIMESTAMPTZ NOT NULL,
    end_date TIMESTAMPTZ,
    auto_renew BOOLEAN NOT NULL DEFAULT false,
    price_kopecks BIGINT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_subscriptions_bathhouse_id ON subscriptions (bathhouse_id);
CREATE INDEX idx_subscriptions_owner_id ON subscriptions (owner_id);
CREATE INDEX idx_subscriptions_status ON subscriptions (status);
CREATE INDEX idx_subscriptions_end_date ON subscriptions (end_date);

-- Promotions table
CREATE TABLE promotions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    bathhouse_id UUID NOT NULL REFERENCES bathhouses(id) ON DELETE CASCADE,
    budget_kopecks BIGINT NOT NULL,
    spent_kopecks BIGINT NOT NULL DEFAULT 0,
    start_date TIMESTAMPTZ NOT NULL,
    end_date TIMESTAMPTZ NOT NULL,
    target_city_id BIGINT REFERENCES cities(id),
    status VARCHAR(30) NOT NULL,
    impression_count BIGINT NOT NULL DEFAULT 0,
    click_count BIGINT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_promotions_bathhouse_id ON promotions (bathhouse_id);
CREATE INDEX idx_promotions_status ON promotions (status);
CREATE INDEX idx_promotions_target_city_id ON promotions (target_city_id);
CREATE INDEX idx_promotions_dates ON promotions (start_date, end_date);

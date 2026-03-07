-- Telegram links
CREATE TABLE telegram_links (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL UNIQUE REFERENCES users(id) ON DELETE CASCADE,
    telegram_id BIGINT NOT NULL UNIQUE,
    telegram_username VARCHAR(255) NOT NULL DEFAULT '',
    linked_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

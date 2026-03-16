-- User preferences for bathhouse recommendations
CREATE TABLE user_preferences (
    user_id UUID PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    preferred_city_id BIGINT,
    price_range_min BIGINT,
    price_range_max BIGINT,
    prefer_pool BOOLEAN NOT NULL DEFAULT false,
    prefer_sauna BOOLEAN NOT NULL DEFAULT false,
    prefer_steam_room BOOLEAN NOT NULL DEFAULT false,
    prefer_hot_tub BOOLEAN NOT NULL DEFAULT false,
    prefer_bbq BOOLEAN NOT NULL DEFAULT false,
    prefer_karaoke BOOLEAN NOT NULL DEFAULT false,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_user_preferences_city ON user_preferences (preferred_city_id);

-- User activity tracking for recommendations and analytics
CREATE TABLE user_activity (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    bathhouse_id UUID NOT NULL REFERENCES bathhouses(id) ON DELETE CASCADE,
    type VARCHAR(20) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_user_activity_user_id ON user_activity (user_id);
CREATE INDEX idx_user_activity_bathhouse_id ON user_activity (bathhouse_id);
CREATE INDEX idx_user_activity_user_created ON user_activity (user_id, created_at);
CREATE INDEX idx_user_activity_type ON user_activity (type);

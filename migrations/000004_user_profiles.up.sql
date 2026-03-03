-- Add profile fields to users table
ALTER TABLE users ADD COLUMN avatar_url TEXT NOT NULL DEFAULT '';
ALTER TABLE users ADD COLUMN bio TEXT NOT NULL DEFAULT '';
ALTER TABLE users ADD COLUMN city_id BIGINT REFERENCES cities(id);

CREATE INDEX idx_users_city_id ON users (city_id);

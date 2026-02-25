-- Enable PostGIS extension for geo-queries
CREATE EXTENSION IF NOT EXISTS postgis;

-- Users table
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email VARCHAR(255) NOT NULL UNIQUE,
    password_hash VARCHAR(255) NOT NULL,
    name VARCHAR(255) NOT NULL,
    phone VARCHAR(50) NOT NULL DEFAULT '',
    role VARCHAR(20) NOT NULL DEFAULT 'client',
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_users_email ON users (email);
CREATE INDEX idx_users_role ON users (role);

-- Cities table
CREATE TABLE cities (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    slug VARCHAR(100) NOT NULL UNIQUE,
    latitude DOUBLE PRECISION NOT NULL DEFAULT 0,
    longitude DOUBLE PRECISION NOT NULL DEFAULT 0
);

CREATE INDEX idx_cities_slug ON cities (slug);

-- Bathhouses table
CREATE TABLE bathhouses (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    owner_id UUID NOT NULL REFERENCES users(id),
    name VARCHAR(255) NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    address VARCHAR(500) NOT NULL,
    city_id BIGINT NOT NULL REFERENCES cities(id),
    latitude DOUBLE PRECISION NOT NULL,
    longitude DOUBLE PRECISION NOT NULL,
    location GEOGRAPHY(Point, 4326),
    price_per_hour BIGINT NOT NULL,
    min_duration INT NOT NULL DEFAULT 1,
    max_guests INT NOT NULL,
    has_pool BOOLEAN NOT NULL DEFAULT false,
    has_sauna BOOLEAN NOT NULL DEFAULT false,
    has_steam_room BOOLEAN NOT NULL DEFAULT false,
    has_hot_tub BOOLEAN NOT NULL DEFAULT false,
    has_bbq BOOLEAN NOT NULL DEFAULT false,
    has_karaoke BOOLEAN NOT NULL DEFAULT false,
    rating DOUBLE PRECISION NOT NULL DEFAULT 0,
    review_count INT NOT NULL DEFAULT 0,
    images JSONB NOT NULL DEFAULT '[]',
    working_hours JSONB NOT NULL DEFAULT '[]',
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Auto-populate location column from lat/lng
CREATE OR REPLACE FUNCTION update_bathhouse_location()
RETURNS TRIGGER AS $$
BEGIN
    NEW.location := ST_SetSRID(ST_MakePoint(NEW.longitude, NEW.latitude), 4326)::geography;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_bathhouse_location
    BEFORE INSERT OR UPDATE OF latitude, longitude ON bathhouses
    FOR EACH ROW EXECUTE FUNCTION update_bathhouse_location();

CREATE INDEX idx_bathhouses_owner_id ON bathhouses (owner_id);
CREATE INDEX idx_bathhouses_city_id ON bathhouses (city_id);
CREATE INDEX idx_bathhouses_status ON bathhouses (status);
CREATE INDEX idx_bathhouses_location ON bathhouses USING GIST (location);

-- Bookings table
CREATE TABLE bookings (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id),
    bathhouse_id UUID NOT NULL REFERENCES bathhouses(id),
    start_time TIMESTAMPTZ NOT NULL,
    end_time TIMESTAMPTZ NOT NULL,
    guest_count INT NOT NULL,
    total_price BIGINT NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    comment TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT chk_booking_times CHECK (end_time > start_time)
);

CREATE INDEX idx_bookings_user_id ON bookings (user_id);
CREATE INDEX idx_bookings_bathhouse_id ON bookings (bathhouse_id);
CREATE INDEX idx_bookings_status ON bookings (status);
CREATE INDEX idx_bookings_time_range ON bookings (bathhouse_id, start_time, end_time);

-- Reviews table
CREATE TABLE reviews (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id),
    bathhouse_id UUID NOT NULL REFERENCES bathhouses(id),
    booking_id UUID NOT NULL REFERENCES bookings(id),
    rating INT NOT NULL CHECK (rating >= 1 AND rating <= 5),
    text TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT uq_reviews_user_booking UNIQUE (user_id, booking_id)
);

CREATE INDEX idx_reviews_bathhouse_id ON reviews (bathhouse_id);
CREATE INDEX idx_reviews_user_id ON reviews (user_id);

-- Representatives table
CREATE TABLE representatives (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id),
    bathhouse_id UUID NOT NULL REFERENCES bathhouses(id),
    owner_id UUID NOT NULL REFERENCES users(id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT uq_representatives_user_bathhouse UNIQUE (user_id, bathhouse_id)
);

CREATE INDEX idx_representatives_user_id ON representatives (user_id);
CREATE INDEX idx_representatives_bathhouse_id ON representatives (bathhouse_id);
CREATE INDEX idx_representatives_owner_id ON representatives (owner_id);

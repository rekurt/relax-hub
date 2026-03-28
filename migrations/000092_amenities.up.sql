CREATE TABLE amenities (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(200) NOT NULL,
    icon VARCHAR(100) NOT NULL DEFAULT '',
    sort_order INT NOT NULL DEFAULT 0,
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Seed with standard amenities matching existing bathhouse boolean fields
INSERT INTO amenities (id, name, icon, sort_order, is_active) VALUES
    (gen_random_uuid(), 'Бассейн', 'pool', 1, true),
    (gen_random_uuid(), 'Сауна', 'sauna', 2, true),
    (gen_random_uuid(), 'Парная', 'steam_room', 3, true),
    (gen_random_uuid(), 'Джакузи', 'hot_tub', 4, true),
    (gen_random_uuid(), 'Мангал', 'bbq', 5, true),
    (gen_random_uuid(), 'Караоке', 'karaoke', 6, true);

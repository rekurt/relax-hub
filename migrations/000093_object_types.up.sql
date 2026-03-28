CREATE TABLE object_types (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(200) NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    sort_order INT NOT NULL DEFAULT 0,
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Seed with standard bathhouse categories
INSERT INTO object_types (id, name, description, sort_order, is_active) VALUES
    (gen_random_uuid(), 'Русская баня', 'Традиционная русская баня с парной', 1, true),
    (gen_random_uuid(), 'Финская сауна', 'Классическая финская сауна с сухим паром', 2, true),
    (gen_random_uuid(), 'Хаммам', 'Турецкая баня с влажным паром', 3, true),
    (gen_random_uuid(), 'Японская баня', 'Офуро и другие японские банные традиции', 4, true),
    (gen_random_uuid(), 'Банный комплекс', 'Комплекс с несколькими видами парных', 5, true),
    (gen_random_uuid(), 'Спа-центр', 'Спа с банными процедурами', 6, true);

CREATE TABLE holidays (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(200) NOT NULL,
    date DATE NOT NULL,
    region VARCHAR(10) NOT NULL DEFAULT 'RU',
    is_recurring BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_holidays_date ON holidays (date);
CREATE INDEX idx_holidays_region ON holidays (region);

CREATE TABLE bathhouse_holiday_prices (
    bathhouse_id UUID NOT NULL REFERENCES bathhouses(id) ON DELETE CASCADE,
    multiplier NUMERIC(3,1) NOT NULL DEFAULT 1.5 CHECK (multiplier >= 1.0 AND multiplier <= 2.0),
    PRIMARY KEY (bathhouse_id)
);

-- Seed initial Russian holidays (all recurring)
INSERT INTO holidays (name, date, region, is_recurring) VALUES
    ('Новый год', '2024-01-01', 'RU', true),
    ('Новогодние каникулы', '2024-01-02', 'RU', true),
    ('Новогодние каникулы', '2024-01-03', 'RU', true),
    ('Новогодние каникулы', '2024-01-04', 'RU', true),
    ('Новогодние каникулы', '2024-01-05', 'RU', true),
    ('Новогодние каникулы', '2024-01-06', 'RU', true),
    ('Рождество', '2024-01-07', 'RU', true),
    ('Новогодние каникулы', '2024-01-08', 'RU', true),
    ('День защитника Отечества', '2024-02-23', 'RU', true),
    ('Международный женский день', '2024-03-08', 'RU', true),
    ('Праздник Весны и Труда', '2024-05-01', 'RU', true),
    ('День Победы', '2024-05-09', 'RU', true),
    ('День России', '2024-06-12', 'RU', true),
    ('День народного единства', '2024-11-04', 'RU', true);

CREATE TABLE seasonal_tariffs (
    id UUID PRIMARY KEY,
    bathhouse_id UUID NOT NULL REFERENCES bathhouses(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    date_from DATE NOT NULL,
    date_to DATE NOT NULL,
    multiplier DECIMAL(5, 2) NOT NULL CHECK (multiplier > 0 AND multiplier <= 10.0),
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT seasonal_tariffs_date_range CHECK (date_to >= date_from)
);

CREATE INDEX idx_seasonal_tariffs_bathhouse ON seasonal_tariffs(bathhouse_id);
CREATE INDEX idx_seasonal_tariffs_active_dates ON seasonal_tariffs(bathhouse_id, is_active, date_from, date_to);

-- Enable pg_trgm extension for fuzzy matching
CREATE EXTENSION IF NOT EXISTS pg_trgm;

-- Add tsvector column for full-text search
ALTER TABLE bathhouses ADD COLUMN search_vector tsvector;

-- Populate search_vector from existing data
UPDATE bathhouses SET search_vector =
    setweight(to_tsvector('russian', coalesce(name, '')), 'A') ||
    setweight(to_tsvector('russian', coalesce(description, '')), 'B') ||
    setweight(to_tsvector('russian', coalesce(address, '')), 'C');

-- GIN index for full-text search
CREATE INDEX idx_bathhouses_search_vector ON bathhouses USING GIN (search_vector);

-- GIN trigram indexes for fuzzy matching
CREATE INDEX idx_bathhouses_name_trgm ON bathhouses USING GIN (name gin_trgm_ops);
CREATE INDEX idx_bathhouses_description_trgm ON bathhouses USING GIN (description gin_trgm_ops);

-- Trigger function to auto-update search_vector on insert/update
CREATE OR REPLACE FUNCTION bathhouses_search_vector_update() RETURNS trigger AS $$
BEGIN
    NEW.search_vector :=
        setweight(to_tsvector('russian', coalesce(NEW.name, '')), 'A') ||
        setweight(to_tsvector('russian', coalesce(NEW.description, '')), 'B') ||
        setweight(to_tsvector('russian', coalesce(NEW.address, '')), 'C');
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER bathhouses_search_vector_trigger
    BEFORE INSERT OR UPDATE OF name, description, address
    ON bathhouses
    FOR EACH ROW
    EXECUTE FUNCTION bathhouses_search_vector_update();

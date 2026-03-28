CREATE TABLE faq (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    category VARCHAR(50) NOT NULL,
    question VARCHAR(500) NOT NULL,
    answer TEXT NOT NULL,
    keywords TEXT[] NOT NULL DEFAULT '{}',
    sort_order INT NOT NULL DEFAULT 0,
    active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_faq_category ON faq(category);
CREATE INDEX idx_faq_active ON faq(active);
CREATE INDEX idx_faq_sort_order ON faq(sort_order);

-- GIN index on keywords array for efficient containment queries
CREATE INDEX idx_faq_keywords ON faq USING GIN(keywords);

-- Trigram index on question for fuzzy text matching
CREATE INDEX idx_faq_question_trgm ON faq USING GIN(question gin_trgm_ops);

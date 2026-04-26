CREATE TABLE grants (
    id            UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    title         TEXT NOT NULL,
    description   TEXT NOT NULL,
    funder        TEXT NOT NULL,
    amount_min    BIGINT,
    amount_max    BIGINT,
    currency      TEXT DEFAULT 'USD',
    deadline      DATE,
    url           TEXT,
    region        TEXT,          -- e.g. "global", "europe", "sub-saharan-africa"
    categories    TEXT[],        -- e.g. ["health","education","climate"]
    eligibility   TEXT,          -- raw eligibility criteria text
    embedding     vector(768),   -- Gemini text-embedding-004 output (768 dims)
    source        TEXT,          -- which scraper populated this
    scraped_at    TIMESTAMPTZ DEFAULT now(),
    created_at    TIMESTAMPTZ DEFAULT now()
);

-- ANN index for fast similarity search
CREATE INDEX grants_embedding_idx ON grants
    USING ivfflat (embedding vector_cosine_ops)
    WITH (lists = 100);

CREATE INDEX grants_deadline_idx ON grants (deadline);
CREATE INDEX grants_categories_idx ON grants USING GIN (categories);

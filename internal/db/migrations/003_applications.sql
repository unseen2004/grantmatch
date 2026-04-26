CREATE TABLE ngo_sessions (
    id            UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name          TEXT NOT NULL,
    mission       TEXT NOT NULL,
    region        TEXT,
    categories    TEXT[],
    budget        BIGINT,
    embedding     vector(768),
    created_at    TIMESTAMPTZ DEFAULT now()
);

CREATE TABLE applications (
    id            UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    session_id    UUID REFERENCES ngo_sessions(id) ON DELETE CASCADE,
    grant_id      UUID REFERENCES grants(id) ON DELETE CASCADE,
    score         FLOAT NOT NULL,        -- 0.0 – 1.0 cosine similarity
    draft_text    TEXT,
    created_at    TIMESTAMPTZ DEFAULT now()
);

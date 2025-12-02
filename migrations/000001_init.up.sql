CREATE TABLE metrics (
    id TEXT PRIMARY KEY,
    mtype TEXT NOT NULL CHECK (mtype IN ('counter', 'gauge')),
    delta BIGINT,
    value DOUBLE PRECISION,
    hash TEXT
);
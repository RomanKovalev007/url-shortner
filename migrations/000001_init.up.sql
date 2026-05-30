CREATE TABLE IF NOT EXISTS urls (
    id         UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    alias      VARCHAR(10) NOT NULL UNIQUE,
    original   TEXT        NOT NULL UNIQUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

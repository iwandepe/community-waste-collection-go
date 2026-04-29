CREATE EXTENSION IF NOT EXISTS "pgcrypto";

CREATE TABLE households (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    owner_name VARCHAR(255) NOT NULL,
    address    TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

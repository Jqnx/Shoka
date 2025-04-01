-- +goose Up
CREATE TABLE IF NOT EXISTS artists (
    id bigserial PRIMARY KEY,
    name text UNIQUE NOT NULL,
    created_at timestamptz NOT NULL DEFAULT NOW(),
    updated_at timestamptz NOT NULL
);

-- +goose Down
DROP TABLE IF EXISTS artists;

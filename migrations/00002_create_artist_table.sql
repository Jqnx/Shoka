-- +goose Up
CREATE TABLE IF NOT EXISTS artists (
    id bigserial PRIMARY KEY,
    name text NOT NULL,
    created_at timestamptz NOT NULL,
    updated_at timestamptz NOT NULL
);

create unique index idx_artist_name on artists(name);

-- +goose Down
DROP TABLE IF EXISTS artists;


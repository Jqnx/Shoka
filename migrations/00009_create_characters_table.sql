-- +goose Up
CREATE TABLE IF NOT EXISTS characters (
    id bigserial PRIMARY KEY,
    name text NOT NULL,
    count bigint NOT NULL
);

create unique index idx_character on characters(name);

-- +goose Down
DROP TABLE IF EXISTS characters;


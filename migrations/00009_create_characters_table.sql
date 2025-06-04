-- +goose Up
CREATE TABLE IF NOT EXISTS characters (
    id bigserial PRIMARY KEY,
    character text NOT NULL,
    count bigint NOT NULL
);

create unique index idx_character on characters(character);

-- +goose Down
DROP TABLE IF EXISTS characters;


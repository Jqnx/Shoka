-- +goose Up
CREATE TABLE IF NOT EXISTS characters (
    id bigserial PRIMARY KEY,
    character text UNIQUE NOT NULL
);

-- +goose Down
DROP TABLE IF EXISTS characters;

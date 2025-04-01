-- +goose Up
CREATE TABLE IF NOT EXISTS groups (
    id bigserial PRIMARY KEY,
    name text UNIQUE NOT NULL
);

-- +goose Down
DROP TABLE IF EXISTS groups;

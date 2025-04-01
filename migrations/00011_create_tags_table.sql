-- +goose Up
CREATE TABLE IF NOT EXISTS tags (
    id bigserial PRIMARY KEY,
    tag text UNIQUE NOT NULL
);

-- +goose Down
DROP TABLE IF EXISTS tags;

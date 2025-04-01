-- +goose Up
CREATE TABLE IF NOT EXISTS parodies (
    id bigserial PRIMARY KEY,
    parody text UNIQUE NOT NULL
);

-- +goose Down
DROP TABLE IF EXISTS parodies;

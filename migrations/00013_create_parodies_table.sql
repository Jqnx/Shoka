-- +goose Up
CREATE TABLE IF NOT EXISTS parodies (
    id bigserial PRIMARY KEY,
    parody text NOT NULL
);

create unique index parody on parodies(parody);

-- +goose Down
DROP TABLE IF EXISTS parodies;


-- +goose Up
CREATE TABLE IF NOT EXISTS parodies (
    id bigserial PRIMARY KEY,
    name text NOT NULL,
    count bigint NOT NULL
);

create unique index idx_parody on parodies(name);

-- +goose Down
DROP TABLE IF EXISTS parodies;


-- +goose Up
CREATE TABLE IF NOT EXISTS tags (
    id bigserial PRIMARY KEY,
    name text NOT NULL,
    count bigint NOT NULL
);

create unique index idx_tag on tags(name);

-- +goose Down
DROP TABLE IF EXISTS tags;


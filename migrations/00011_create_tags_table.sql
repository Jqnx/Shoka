-- +goose Up
CREATE TABLE IF NOT EXISTS tags (
    id bigserial PRIMARY KEY,
    tag text NOT NULL
);

create unique index idx_tag on tags(tag);

-- +goose Down
DROP TABLE IF EXISTS tags;


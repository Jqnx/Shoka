-- +goose Up
CREATE TABLE IF NOT EXISTS groups (
    id bigserial PRIMARY KEY,
    name text NOT NULL,
    created_at timestamptz NOT NULL DEFAULT NOW(),
    updated_at timestamptz NOT NULL
);

create unique index idx_group_name on groups(name);

-- +goose Down
DROP TABLE IF EXISTS groups;


-- +goose Up
CREATE TABLE IF NOT EXISTS archives (
    id bigserial PRIMARY KEY,
    title text UNIQUE NOT NULL,
    summary text,
    lang varchar(2),
    category varchar(10),
    page_count bigint NOT NULL DEFAULT 0,
    file_path text UNIQUE,
    a_id bigint NOT NULL DEFAULT 1,
    created_at timestamptz NOT NULL DEFAULT NOW(),
    updated_at timestamptz NOT NULL
);

create unique index idx_a_id on archives(a_id);

-- +goose Down
DROP TABLE IF EXISTS archives;


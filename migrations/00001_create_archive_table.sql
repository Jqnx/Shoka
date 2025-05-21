-- +goose Up
CREATE TABLE IF NOT EXISTS archives (
    id bigserial PRIMARY KEY,
    title text NOT NULL,
    summary text,
    language varchar(2),
    category text,
    page_count bigint NOT NULL DEFAULT 0,
    file_path text UNIQUE,
    archive_id text NOT NULL,
    hash text,
    thumbs_path text,
    cover_path text,
-- pages_path text,
    type text NOT NULL,
    created_at timestamptz NOT NULL DEFAULT NOW(),
    updated_at timestamptz NOT NULL
);

create unique index idx_archive_id on archives(archive_id);
create unique index idx_thumbs_path on archives(thumbs_path);
create unique index idx_cover_path on archives(cover_path);
create unique index idx_hash on archives(hash);
create index idx_title on archives(title);

-- +goose Down
DROP TABLE IF EXISTS archives;


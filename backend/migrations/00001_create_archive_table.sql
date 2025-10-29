-- +goose Up
create table if not exists archives (
    id char(8) primary key,
    title text not null,
    summary text,
    language char(2),
    category text,
    page_count smallint not null default 0,
    file_path text unique not null,
    file_name text unique not null,
    hash text not null,
    thumbs_path text,
    type text not null,
    created_at timestamptz not null,
    updated_at timestamptz not null,
    release_date timestamptz,
    search_vector tsvector
);

-- create unique index idx_archive_id on archives(archive_id);
create unique index idx_thumbs_path on archives (thumbs_path);
create unique index idx_hash on archives (hash);
create index idx_title on archives (title);
create index idx_search_vector on archives using gin (search_vector);

-- +goose Down
drop table if exists archives;


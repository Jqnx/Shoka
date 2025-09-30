-- +goose Up
create table if not exists artists (
    id bigserial primary key,
    name text not null,
    count int not null,
    created_at timestamptz not null,
    updated_at timestamptz not null
);

create unique index idx_artist_name on artists (name);

-- +goose Down
drop table if exists artists;


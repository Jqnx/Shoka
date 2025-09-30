-- +goose Up
create table if not exists artist_aliases (
    id bigserial primary key,
    alias text unique not null,
    artist_id bigint not null,
    foreign key (artist_id) references artists (id)
);

create index idx_alias_artist_id on artist_aliases (artist_id);

-- +goose Down
drop table if exists artist_aliases;


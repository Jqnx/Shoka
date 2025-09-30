-- +goose Up
create table if not exists artist_links (
    id bigserial primary key,
    link text unique not null,
    artist_id bigint not null,
    foreign key (artist_id) references artists (id)
);

create index idx_links_artist_id on artist_links (artist_id);

-- +goose Down
drop table if exists artist_links;


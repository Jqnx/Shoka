-- +goose Up
create table if not exists characters (
    id bigserial primary key,
    name text not null,
    count bigint not null
);

create unique index idx_character on characters (name);

-- +goose Down
drop table if exists characters;


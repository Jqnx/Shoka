-- +goose Up
create table if not exists tags (
    id bigserial primary key,
    name text not null,
    count bigint not null
);

create unique index idx_tag on tags (name);

-- +goose Down
drop table if exists tags;


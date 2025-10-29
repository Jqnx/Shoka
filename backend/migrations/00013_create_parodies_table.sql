-- +goose Up
create table if not exists parodies (
    id bigserial primary key,
    name text not null,
    count bigint not null
);

create unique index idx_parody on parodies (name);

-- +goose Down
drop table if exists parodies;


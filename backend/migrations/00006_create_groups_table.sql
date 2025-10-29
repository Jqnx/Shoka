-- +goose Up
create table if not exists groups (
    id bigserial primary key,
    name text not null,
    created_at timestamptz not null,
    updated_at timestamptz not null
);

create unique index idx_group_name on groups (name);

-- +goose Down
drop table if exists groups;


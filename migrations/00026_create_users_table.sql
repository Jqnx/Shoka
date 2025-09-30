-- +goose Up
-- +goose StatementBegin
create table if not exists users (
    id uuid primary key,
    name text not null unique,
    password text not null,
    created_at timestamptz not null,
    updated_at timestamptz not null
);
-- +goose StatementEnd
-- +goose Down
-- +goose StatementBegin
drop table if exists users;
-- +goose StatementEnd



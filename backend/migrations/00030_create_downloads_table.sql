-- +goose Up
-- +goose StatementBegin
create table if not exists downloads (
    id uuid primary key,
    url text not null,
    source text not null,
    filename text not null,
    status text not null,
    progress int default 0,
    error text,
    created_at timestamptz not null,
    updated_at timestamptz not null,
    speed bigint,
    total_size bigint,
    downloaded bigint,
    started_at timestamptz,
    can_resume boolean,
    resume_supported boolean
);

create index if not exists idx_downloads_status on downloads (status);
-- +goose StatementEnd
-- +goose Down
-- +goose StatementBegin
drop table if exists downloads;
-- +goose StatementEnd



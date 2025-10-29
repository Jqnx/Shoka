-- +goose Up
create table if not exists urls (
    id bigserial primary key,
    url text not null,
    archive_id char(8) not null,
    foreign key (archive_id) references archives (id)
);

create index idx_url_archive_id on urls(archive_id);

-- +goose Down
drop table if exists urls;


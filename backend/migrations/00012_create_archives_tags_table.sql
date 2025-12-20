-- +goose Up
create table if not exists archives_tags (
    archive_id char(8),
    tag_id bigint,
    primary key (archive_id, tag_id),
    foreign key (archive_id) references archives (id) on delete cascade,
    foreign key (tag_id) references tags (id) on delete cascade
);

-- +goose Down
drop table if exists archives_tags;


-- +goose Up
create table if not exists archives_parodies (
    archive_id char(8),
    parody_id bigint,
    primary key (archive_id, parody_id),
    foreign key (archive_id) references archives (id) on delete cascade,
    foreign key (parody_id) references parodies (id) on delete cascade
);

-- +goose Down
drop table if exists archives_parodies;


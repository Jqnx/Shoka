-- +goose Up
create table if not exists archives_parodies (
    archive_id char(8),
    parody_id bigint,
    primary key (archive_id, parody_id),
    foreign key (archive_id) references archives (id),
    foreign key (parody_id) references parodies (id)
);

-- +goose Down
drop table if exists archives_parodies;


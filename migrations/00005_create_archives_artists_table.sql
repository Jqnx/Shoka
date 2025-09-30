-- +goose Up
create table if not exists archives_artists (
    archive_id char(8),
    artist_id bigint,
    primary key (archive_id, artist_id),
    foreign key (archive_id) references archives (id),
    foreign key (artist_id) references artists (id)
);

-- +goose Down
drop table if exists archives_artists;


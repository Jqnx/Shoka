-- +goose Up
create table if not exists artists_groups (
    artist_id bigint,
    group_id bigint,
    primary key (artist_id, group_id),
    foreign key (artist_id) references artists (id),
    foreign key (group_id) references groups (id)
);

-- +goose Down
drop table if exists artists_groups;


-- +goose Up
create table if not exists archives_characters (
    archive_id char(8),
    character_id bigint,
    primary key (archive_id, character_id),
    foreign key (archive_id) references archives (id) on delete cascade,
    foreign key (character_id) references characters (id) on delete cascade
);

-- +goose Down
drop table if exists archives_characters;


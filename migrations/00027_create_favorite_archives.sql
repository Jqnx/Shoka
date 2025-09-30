-- +goose Up
-- +goose StatementBegin
create table if not exists favorite_archives (
    archive_id char(8),
    user_id uuid,
    favorited_at timestamptz not null,
    primary key (archive_id, user_id),
    foreign key (archive_id) references archives (id),
    foreign key (user_id) references users (id)
);
-- +goose StatementEnd
-- +goose Down
-- +goose StatementBegin
drop table if exists favorite_archives;
-- +goose StatementEnd



-- +goose Up
-- +goose StatementBegin
create table if not exists reading_progress (
    archive_id char(8),
    user_id uuid,
    page smallint not null,
    state text not null,
    last_read timestamptz not null,
    primary key (archive_id, user_id),
    foreign key (archive_id) references archives (id),
    foreign key (user_id) references users (id)
);
-- +goose StatementEnd
-- +goose Down
-- +goose StatementBegin
drop table if exists reading_progress;
-- +goose StatementEnd



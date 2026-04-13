-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS reading_progress (
    archive_id TEXT,
    user_id TEXT,
    page INTEGER NOT NULL,
    status TEXT NOT NULL,
    last_read DATETIME NOT NULL,
    PRIMARY KEY (archive_id, user_id),
    FOREIGN KEY (archive_id) REFERENCES archive (id) ON DELETE CASCADE,
    FOREIGN KEY (user_id) REFERENCES user (id)
);
-- +goose StatementEnd
-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS reading_progress;
-- +goose StatementEnd

-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS progress (
    archive_id TEXT NOT NULL,
    user_id TEXT NOT NULL,
    page INTEGER NOT NULL DEFAULT 0,
    completed BOOLEAN NOT NULL DEFAULT FALSE,
    last_read DATETIME NOT NULL DEFAULT (datetime('now')),
    PRIMARY KEY (user_id, archive_id),
    FOREIGN KEY (archive_id) REFERENCES archive (id) ON DELETE CASCADE,
    FOREIGN KEY (user_id) REFERENCES user (id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_progress_user_archive ON progress(user_id, archive_id);
-- +goose StatementEnd
-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS reading_progress;
-- +goose StatementEnd

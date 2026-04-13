-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS favorite_archive (
    archive_id TEXT,
    user_id TEXT,
    favorited_at DATETIME NOT NULL,
    PRIMARY KEY (archive_id, user_id),
    FOREIGN KEY (archive_id) REFERENCES archive (id) ON DELETE CASCADE,
    FOREIGN KEY (user_id) REFERENCES user (id)
);
-- +goose StatementEnd
-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS favorite_archive;
-- +goose StatementEnd

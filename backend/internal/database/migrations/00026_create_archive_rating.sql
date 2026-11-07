-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS archive_rating (
    archive_id TEXT NOT NULL,
    user_id TEXT NOT NULL,
    rating INTEGER NOT NULL CHECK (rating BETWEEN 1 AND 5),
    rated_at DATETIME NOT NULL,
    PRIMARY KEY (archive_id, user_id),
    FOREIGN KEY (archive_id) REFERENCES archive (id) ON DELETE CASCADE,
    FOREIGN KEY (user_id) REFERENCES user (id) ON DELETE CASCADE
);
-- +goose StatementEnd
-- +goose StatementBegin
CREATE INDEX IF NOT EXISTS idx_archive_rating_user ON archive_rating (user_id);
-- +goose StatementEnd
-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_archive_rating_user;
-- +goose StatementEnd
-- +goose StatementBegin
DROP TABLE IF EXISTS archive_rating;
-- +goose StatementEnd

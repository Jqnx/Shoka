-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS favorite_page (
    archive_id TEXT NOT NULL,
    user_id TEXT NOT NULL,
    page_index INTEGER NOT NULL CHECK (page_index >= 0),
    favorited_at DATETIME NOT NULL,
    PRIMARY KEY (archive_id, user_id, page_index),
    FOREIGN KEY (archive_id) REFERENCES archive (id) ON DELETE CASCADE,
    FOREIGN KEY (user_id) REFERENCES user (id) ON DELETE CASCADE
);
-- +goose StatementEnd
-- +goose StatementBegin
CREATE INDEX IF NOT EXISTS idx_favorite_page_user ON favorite_page (user_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_favorite_page_user;
-- +goose StatementEnd
-- +goose StatementBegin
DROP TABLE IF EXISTS favorite_page;
-- +goose StatementEnd

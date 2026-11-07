-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS library (
    id         TEXT PRIMARY KEY NOT NULL,
    name       TEXT NOT NULL,
    path       TEXT NOT NULL UNIQUE,
    type       TEXT NOT NULL DEFAULT 'doujinshi'
                       CHECK (type IN ('doujinshi', 'audio')),
    enabled    INTEGER NOT NULL DEFAULT 1,
    created_at DATETIME NOT NULL DEFAULT (datetime('now')),
    updated_at DATETIME NOT NULL DEFAULT (datetime('now'))
);
-- +goose StatementEnd
-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS library;
-- +goose StatementEnd

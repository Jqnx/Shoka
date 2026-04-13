-- +goose Up
CREATE TABLE IF NOT EXISTS tag (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL UNIQUE,
    description TEXT,
    count INTEGER NOT NULL
);

-- +goose Down
DROP TABLE IF EXISTS tag;

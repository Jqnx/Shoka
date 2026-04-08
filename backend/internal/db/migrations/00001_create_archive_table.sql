-- +goose Up
CREATE TABLE IF NOT EXISTS archive (
    id TEXT PRIMARY KEY NOT NULL,
    title TEXT NOT NULL,
    summary TEXT,
    language TEXT,
    category TEXT,
    page_count INTEGER NOT NULL,
    file_path TEXT UNIQUE NOT NULL,
    file_hash TEXT UNIQUE NOT NULL,
    created_at DATETIME NOT NULL,
    updated_at DATETIME NOT NULL,
    release_date DATETIME
);

create index idx_title ON archive (title);

-- +goose Down
DROP TABLE IF EXISTS archive;

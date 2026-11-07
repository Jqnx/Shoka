-- +goose Up
CREATE TABLE IF NOT EXISTS archive (
    id TEXT PRIMARY KEY NOT NULL,
    title TEXT NOT NULL,
    summary TEXT,
    language TEXT,
    category TEXT,
    page_count INTEGER NOT NULL default 0,
    file_path TEXT UNIQUE NOT NULL,
    file_size INTEGER NOT NULL,
    mod_time DATETIME NOT NULL,
    created_at DATETIME NOT NULL default (datetime('now')),
    updated_at DATETIME NOT NULL default (datetime('now')),
    release_date DATETIME
);

create index idx_title ON archive (title);

-- +goose Down
DROP TABLE IF EXISTS archive;

-- +goose Up
-- archive_url has never been written to (migration 00008's table and its sqlc
-- queries were dead code), so a clean drop/recreate is safe. SQLite can't add
-- a table-level PRIMARY KEY/UNIQUE via ALTER anyway.
DROP TABLE IF EXISTS archive_url;

CREATE TABLE archive_url (
    archive_id TEXT NOT NULL,
    url        TEXT NOT NULL,
    PRIMARY KEY (archive_id, url),
    FOREIGN KEY (archive_id) REFERENCES archive (id) ON DELETE CASCADE
);

-- +goose Down
DROP TABLE IF EXISTS archive_url;

CREATE TABLE archive_url (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    url TEXT NOT NULL,
    archive_id TEXT NOT NULL,
    FOREIGN KEY (archive_id) REFERENCES archive (id) ON DELETE CASCADE
);

CREATE INDEX idx_url_archive_id ON archive_url (archive_id);

-- +goose Up
CREATE TABLE IF NOT EXISTS archive_url (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    url TEXT NOT NULL,
    archive_id TEXT NOT NULL,
    FOREIGN KEY (archive_id) REFERENCES archive (id) ON DELETE CASCADE
);

CREATE INDEX idx_url_archive_id ON archive_url (archive_id);

-- +goose Down
DROP TABLE IF EXISTS archive_url;

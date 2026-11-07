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
-- Unlike the Up block's drop, this table may hold real rows by the time
-- anyone rolls back (every archive's source links, once this feature has
-- shipped), so rename-and-copy instead of dropping them outright.
ALTER TABLE archive_url RENAME TO archive_url_new;

CREATE TABLE archive_url (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    url TEXT NOT NULL,
    archive_id TEXT NOT NULL,
    FOREIGN KEY (archive_id) REFERENCES archive (id) ON DELETE CASCADE
);

CREATE INDEX idx_url_archive_id ON archive_url (archive_id);

INSERT INTO archive_url (url, archive_id)
SELECT url, archive_id FROM archive_url_new;

DROP TABLE archive_url_new;

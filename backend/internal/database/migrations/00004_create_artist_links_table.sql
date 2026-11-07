-- +goose Up
CREATE TABLE IF NOT EXISTS artist_url (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    artist_id INTEGER NOT NULL,
    url TEXT UNIQUE NOT NULL,
    FOREIGN KEY (artist_id) REFERENCES artist (id) ON DELETE CASCADE
);

CREATE INDEX idx_url_artist_id ON artist_url (artist_id);

-- +goose Down
DROP TABLE IF EXISTS artist_url;

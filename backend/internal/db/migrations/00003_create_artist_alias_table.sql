-- +goose Up
CREATE TABLE IF NOT EXISTS artist_alias (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    artist_id INTEGER NOT NULL,
    alias TEXT UNIQUE NOT NULL,
    FOREIGN KEY (artist_id) REFERENCES artist (id) ON DELETE CASCADE
);

CREATE INDEX idx_alias_artist_id ON artist_alias (artist_id);

-- +goose Down
DROP TABLE IF EXISTS artist_alias;

-- +goose Up
CREATE TABLE IF NOT EXISTS archive_artist (
    archive_id TEXT,
    artist_id INTEGER,
    PRIMARY KEY (archive_id, artist_id),
    FOREIGN KEY (archive_id) REFERENCES archive (id) ON DELETE CASCADE,
    FOREIGN KEY (artist_id) REFERENCES artist (id) ON DELETE CASCADE
);

-- +goose Down
DROP TABLE IF EXISTS archive_artist;

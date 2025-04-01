-- +goose Up
CREATE TABLE IF NOT EXISTS archives_artists (
    archive_id bigint,
    artist_id bigint,
    PRIMARY KEY (archive_id, artist_id),
    FOREIGN KEY (archive_id) REFERENCES archives (id),
    FOREIGN KEY (artist_id) REFERENCES artists (id)
);

-- +goose Down
DROP TABLE IF EXISTS archives_artists;

-- +goose Up
CREATE TABLE IF NOT EXISTS archive_parody (
    archive_id TEXT,
    parody_id INTEGER,
    PRIMARY KEY (archive_id, parody_id),
    FOREIGN KEY (archive_id) REFERENCES archive (id) ON DELETE CASCADE,
    FOREIGN KEY (parody_id) REFERENCES parody (id) ON DELETE CASCADE
);

-- +goose Down
DROP TABLE IF EXISTS archive_parody;

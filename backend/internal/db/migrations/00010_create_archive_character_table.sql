-- +goose Up
CREATE TABLE IF NOT EXISTS archive_character (
    archive_id TEXT,
    character_id INTEGER,
    PRIMARY KEY (archive_id, character_id),
    FOREIGN KEY (archive_id) REFERENCES archive (id) ON DELETE CASCADE,
    FOREIGN KEY (character_id) REFERENCES character (id) ON DELETE CASCADE
);

-- +goose Down
DROP TABLE IF EXISTS archive_character;

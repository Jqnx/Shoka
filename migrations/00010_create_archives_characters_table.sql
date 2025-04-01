-- +goose Up
CREATE TABLE IF NOT EXISTS archives_characters (
    archive_id bigint,
    character_id bigint,
    PRIMARY KEY (archive_id, character_id),
    FOREIGN KEY (archive_id) REFERENCES archives (id),
    FOREIGN KEY (character_id) REFERENCES characters (id)
);

-- +goose Down
DROP TABLE IF EXISTS archives_characters;

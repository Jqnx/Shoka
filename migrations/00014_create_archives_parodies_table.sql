-- +goose Up
CREATE TABLE IF NOT EXISTS archives_parodies (
    archive_id bigint,
    parody_id bigint,
    PRIMARY KEY (archive_id, parody_id),
    FOREIGN KEY (archive_id) REFERENCES archives (id),
    FOREIGN KEY (parody_id) REFERENCES parodies (id)
);

-- +goose Down
DROP TABLE IF EXISTS archives_parodies;

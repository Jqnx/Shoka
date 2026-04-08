-- +goose Up
CREATE TABLE IF NOT EXISTS archive_tag (
    archive_id TEXT,
    tag_id INTEGER,
    PRIMARY KEY (archive_id, tag_id),
    FOREIGN KEY (archive_id) REFERENCES archive (id) ON DELETE CASCADE,
    FOREIGN KEY (tag_id) REFERENCES tag (id) ON DELETE CASCADE
);

-- +goose Down
DROP TABLE IF EXISTS archive_tag;

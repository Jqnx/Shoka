-- +goose Up
CREATE TABLE IF NOT EXISTS archives_tags (
    archive_id bigint,
    tag_id bigint,
    PRIMARY KEY (archive_id, tag_id),
    FOREIGN KEY (archive_id) REFERENCES archives (id),
    FOREIGN KEY (tag_id) REFERENCES tags (id)
);

-- +goose Down
DROP TABLE IF EXISTS archives_tags;

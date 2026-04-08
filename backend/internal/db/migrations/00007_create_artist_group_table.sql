-- +goose Up
CREATE TABLE IF NOT EXISTS artist_group (
    artist_id INTEGER,
    group_id INTEGER,
    PRIMARY KEY (artist_id, group_id),
    FOREIGN KEY (artist_id) REFERENCES artist (id) ON DELETE CASCADE,
    FOREIGN KEY (group_id) REFERENCES group (id) ON DELETE CASCADE
);

-- +goose Down
DROP TABLE IF EXISTS artist_group;

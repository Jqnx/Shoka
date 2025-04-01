-- +goose Up
CREATE TABLE IF NOT EXISTS artists_groups (
    artist_id bigint,
    group_id bigint,
    PRIMARY KEY (artist_id, group_id),
    FOREIGN KEY (artist_id) REFERENCES artists (id),
    FOREIGN KEY (group_id) REFERENCES groups (id)
);

-- +goose Down
DROP TABLE IF EXISTS artists_groups;

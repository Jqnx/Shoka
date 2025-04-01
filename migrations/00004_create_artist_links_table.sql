-- +goose Up
CREATE TABLE IF NOT EXISTS artist_links (
    id bigserial PRIMARY KEY,
    link text UNIQUE NOT NULL,
    artist_id bigint NOT NULL,
    FOREIGN KEY (artist_id) REFERENCES artists (id)
);

-- +goose Down
DROP TABLE IF EXISTS artist_links;

-- +goose Up
CREATE TABLE IF NOT EXISTS artist_links (
    id bigserial PRIMARY KEY,
    link text UNIQUE NOT NULL,
    artist_id bigint NOT NULL,
    FOREIGN KEY (artist_id) REFERENCES artists (id)
);

create index idx_links_artist_id on artist_links(artist_id);

-- +goose Down
DROP TABLE IF EXISTS artist_links;


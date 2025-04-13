-- +goose Up
CREATE TABLE IF NOT EXISTS artist_aliases (
    id bigserial PRIMARY KEY,
    alias text UNIQUE NOT NULL,
    artist_id bigint NOT NULL,
    FOREIGN KEY (artist_id) REFERENCES artists (id)
);

create index idx_alias_artist_id on artist_aliases(artist_id);

-- +goose Down
DROP TABLE IF EXISTS artist_aliases;


-- +goose Up
-- GetCircleArtists filters on circle_id, the trailing column of the
-- artist_circle primary key (artist_id, circle_id), so without this index
-- SQLite scans the whole table for every circle page.
CREATE INDEX IF NOT EXISTS idx_artist_circle_circle ON artist_circle (circle_id);

-- +goose Down
DROP INDEX IF EXISTS idx_artist_circle_circle;

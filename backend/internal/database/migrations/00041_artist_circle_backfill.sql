-- +goose Up
INSERT OR IGNORE INTO artist_circle (artist_id, circle_id)
SELECT aa.artist_id, ac.circle_id
FROM archive_artist aa
JOIN archive_circle ac ON ac.archive_id = aa.archive_id;

-- +goose Down
-- The table had no writer before this migration, so every row is ours.
DELETE FROM artist_circle;

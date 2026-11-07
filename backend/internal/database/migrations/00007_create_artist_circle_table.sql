-- +goose Up
CREATE TABLE IF NOT EXISTS artist_circle (
    artist_id INTEGER,
    circle_id INTEGER,
    PRIMARY KEY (artist_id, circle_id),
    FOREIGN KEY (artist_id) REFERENCES artist (id) ON DELETE CASCADE,
    FOREIGN KEY (circle_id) REFERENCES circle (id) ON DELETE CASCADE
);

-- +goose Down
DROP TABLE IF EXISTS artist_circle;

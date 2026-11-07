-- +goose Up
-- An artist's thumbnail is one of its archives' covers. NULL means "the
-- artist's oldest archive" (see GetArtistList), so the image follows the
-- library without a stored id. ON DELETE SET NULL clears a deleted choice.
ALTER TABLE artist ADD COLUMN cover_archive_id TEXT REFERENCES archive (id) ON DELETE SET NULL;

-- archive_artist's primary key is (archive_id, artist_id); every
-- "archives by artist" lookup and the cover resolution above scan by artist.
CREATE INDEX IF NOT EXISTS idx_archive_artist_artist ON archive_artist (artist_id);

CREATE TABLE IF NOT EXISTS artist_favorite (
    artist_id INTEGER NOT NULL,
    user_id TEXT NOT NULL,
    favorited_at DATETIME NOT NULL,
    PRIMARY KEY (artist_id, user_id),
    FOREIGN KEY (artist_id) REFERENCES artist (id) ON DELETE CASCADE,
    FOREIGN KEY (user_id) REFERENCES user (id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_artist_favorite_user ON artist_favorite (user_id);

CREATE TABLE IF NOT EXISTS artist_rating (
    artist_id INTEGER NOT NULL,
    user_id TEXT NOT NULL,
    rating INTEGER NOT NULL CHECK (rating BETWEEN 1 AND 5),
    rated_at DATETIME NOT NULL,
    PRIMARY KEY (artist_id, user_id),
    FOREIGN KEY (artist_id) REFERENCES artist (id) ON DELETE CASCADE,
    FOREIGN KEY (user_id) REFERENCES user (id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_artist_rating_user ON artist_rating (user_id);

-- +goose Down
DROP INDEX IF EXISTS idx_artist_rating_user;
DROP TABLE IF EXISTS artist_rating;
DROP INDEX IF EXISTS idx_artist_favorite_user;
DROP TABLE IF EXISTS artist_favorite;
DROP INDEX IF EXISTS idx_archive_artist_artist;
ALTER TABLE artist DROP COLUMN cover_archive_id;

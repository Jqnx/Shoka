-- +goose Up
-- +goose StatementBegin
-- Per-user favorites of individual gallery images of an audio work. Keyed by
-- the 0-based gallery index (the same index the API and UI use), so it
-- mirrors favorite_page's archive/page pair.
CREATE TABLE IF NOT EXISTS audio_image_favorite (
    audio_id TEXT NOT NULL,
    user_id TEXT NOT NULL,
    image_index INTEGER NOT NULL CHECK (image_index >= 0),
    favorited_at DATETIME NOT NULL,
    PRIMARY KEY (audio_id, user_id, image_index),
    FOREIGN KEY (audio_id) REFERENCES audio (id) ON DELETE CASCADE,
    FOREIGN KEY (user_id) REFERENCES user (id) ON DELETE CASCADE
);
-- +goose StatementEnd
-- +goose StatementBegin
CREATE INDEX IF NOT EXISTS idx_audio_image_favorite_user ON audio_image_favorite (user_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_audio_image_favorite_user;
-- +goose StatementEnd
-- +goose StatementBegin
DROP TABLE IF EXISTS audio_image_favorite;
-- +goose StatementEnd

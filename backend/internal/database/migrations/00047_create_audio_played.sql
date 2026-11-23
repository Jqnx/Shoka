-- +goose Up
-- +goose StatementBegin
CREATE TABLE audio_played (
    audio_id TEXT NOT NULL REFERENCES audio (id) ON DELETE CASCADE,
    user_id TEXT NOT NULL REFERENCES user (id) ON DELETE CASCADE,
    played_at DATETIME NOT NULL,
    PRIMARY KEY (audio_id, user_id)
);
-- +goose StatementEnd

CREATE INDEX idx_audio_played_user ON audio_played (user_id);

-- +goose Down
DROP INDEX IF EXISTS idx_audio_played_user;
-- +goose StatementBegin
DROP TABLE IF EXISTS audio_played;
-- +goose StatementEnd

-- +goose Up
-- +goose StatementBegin
CREATE TABLE audio (
    id TEXT PRIMARY KEY NOT NULL,
    library_id TEXT NOT NULL REFERENCES library (id) ON DELETE CASCADE,
    title TEXT NOT NULL,
    folder_path TEXT UNIQUE NOT NULL,
    mod_time DATETIME NOT NULL,
    duration_seconds INTEGER NOT NULL DEFAULT 0,
    release_date DATETIME,
    language TEXT,
    created_at DATETIME NOT NULL DEFAULT (datetime('now')),
    updated_at DATETIME NOT NULL DEFAULT (datetime('now'))
);
-- +goose StatementEnd

CREATE INDEX idx_audio_library ON audio (library_id);
CREATE INDEX idx_audio_title ON audio (title);

-- +goose StatementBegin
CREATE TABLE audio_file (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    audio_id TEXT NOT NULL REFERENCES audio (id) ON DELETE CASCADE,
    relative_path TEXT NOT NULL,
    duration_seconds INTEGER NOT NULL DEFAULT 0,
    sort_order INTEGER NOT NULL,
    UNIQUE (audio_id, relative_path)
);
-- +goose StatementEnd

-- +goose StatementBegin
CREATE TABLE audio_image (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    audio_id TEXT NOT NULL REFERENCES audio (id) ON DELETE CASCADE,
    relative_path TEXT NOT NULL,
    sort_order INTEGER NOT NULL,
    UNIQUE (audio_id, relative_path)
);
-- +goose StatementEnd

-- +goose StatementBegin
CREATE TABLE voice_actor (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT UNIQUE NOT NULL,
    count INTEGER NOT NULL DEFAULT 0
);
-- +goose StatementEnd

-- +goose StatementBegin
CREATE TABLE audio_artist (
    audio_id TEXT NOT NULL REFERENCES audio (id) ON DELETE CASCADE,
    artist_id INTEGER NOT NULL REFERENCES artist (id) ON DELETE CASCADE,
    PRIMARY KEY (audio_id, artist_id)
);
-- +goose StatementEnd

-- +goose StatementBegin
CREATE TABLE audio_tag (
    audio_id TEXT NOT NULL REFERENCES audio (id) ON DELETE CASCADE,
    tag_id INTEGER NOT NULL REFERENCES tag (id) ON DELETE CASCADE,
    PRIMARY KEY (audio_id, tag_id)
);
-- +goose StatementEnd

-- +goose StatementBegin
CREATE TABLE audio_circle (
    audio_id TEXT NOT NULL REFERENCES audio (id) ON DELETE CASCADE,
    circle_id INTEGER NOT NULL REFERENCES circle (id) ON DELETE CASCADE,
    PRIMARY KEY (audio_id, circle_id)
);
-- +goose StatementEnd

-- +goose StatementBegin
CREATE TABLE audio_voice_actor (
    audio_id TEXT NOT NULL REFERENCES audio (id) ON DELETE CASCADE,
    voice_actor_id INTEGER NOT NULL REFERENCES voice_actor (id) ON DELETE CASCADE,
    PRIMARY KEY (audio_id, voice_actor_id)
);
-- +goose StatementEnd

-- +goose StatementBegin
CREATE TABLE audio_favorite (
    audio_id TEXT NOT NULL REFERENCES audio (id) ON DELETE CASCADE,
    user_id TEXT NOT NULL REFERENCES user (id) ON DELETE CASCADE,
    favorited_at DATETIME NOT NULL,
    PRIMARY KEY (audio_id, user_id)
);
-- +goose StatementEnd

-- +goose StatementBegin
CREATE TABLE audio_rating (
    audio_id TEXT NOT NULL REFERENCES audio (id) ON DELETE CASCADE,
    user_id TEXT NOT NULL REFERENCES user (id) ON DELETE CASCADE,
    rating INTEGER NOT NULL CHECK (rating BETWEEN 1 AND 5),
    rated_at DATETIME NOT NULL,
    PRIMARY KEY (audio_id, user_id)
);
-- +goose StatementEnd

CREATE INDEX idx_audio_artist_artist ON audio_artist (artist_id);
CREATE INDEX idx_audio_tag_tag ON audio_tag (tag_id);
CREATE INDEX idx_audio_circle_circle ON audio_circle (circle_id);
CREATE INDEX idx_audio_voice_actor_actor ON audio_voice_actor (voice_actor_id);
CREATE INDEX idx_audio_favorite_user ON audio_favorite (user_id);
CREATE INDEX idx_audio_rating_user ON audio_rating (user_id);

-- +goose Down
DROP INDEX IF EXISTS idx_audio_rating_user;
DROP INDEX IF EXISTS idx_audio_favorite_user;
DROP INDEX IF EXISTS idx_audio_voice_actor_actor;
DROP INDEX IF EXISTS idx_audio_circle_circle;
DROP INDEX IF EXISTS idx_audio_tag_tag;
DROP INDEX IF EXISTS idx_audio_artist_artist;
-- +goose StatementBegin
DROP TABLE IF EXISTS audio_rating;
-- +goose StatementEnd
-- +goose StatementBegin
DROP TABLE IF EXISTS audio_favorite;
-- +goose StatementEnd
-- +goose StatementBegin
DROP TABLE IF EXISTS audio_voice_actor;
-- +goose StatementEnd
-- +goose StatementBegin
DROP TABLE IF EXISTS audio_circle;
-- +goose StatementEnd
-- +goose StatementBegin
DROP TABLE IF EXISTS audio_tag;
-- +goose StatementEnd
-- +goose StatementBegin
DROP TABLE IF EXISTS audio_artist;
-- +goose StatementEnd
-- +goose StatementBegin
DROP TABLE IF EXISTS voice_actor;
-- +goose StatementEnd
-- +goose StatementBegin
DROP TABLE IF EXISTS audio_image;
-- +goose StatementEnd
-- +goose StatementBegin
DROP TABLE IF EXISTS audio_file;
-- +goose StatementEnd
-- +goose StatementBegin
DROP TABLE IF EXISTS audio;
-- +goose StatementEnd
DROP INDEX IF EXISTS idx_audio_title;
DROP INDEX IF EXISTS idx_audio_library;

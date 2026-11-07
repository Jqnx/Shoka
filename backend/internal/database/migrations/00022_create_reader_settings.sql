-- +goose Up
-- +goose StatementBegin
-- One row per user, created lazily on first PATCH - a missing row just means
-- "using defaults" (see ReaderSettingsHandler), so this intentionally isn't
-- seeded for every existing user.
CREATE TABLE IF NOT EXISTS reader_settings (
    user_id TEXT PRIMARY KEY,
    -- 'ltr' | 'rtl' | 'vertical' (webtoon-style continuous scroll)
    reading_direction TEXT NOT NULL DEFAULT 'rtl',
    -- 'single' | 'double' (two-page spread)
    page_layout TEXT NOT NULL DEFAULT 'single',
    -- 'width' | 'height' | 'original'
    fit_mode TEXT NOT NULL DEFAULT 'width',
    -- 'black' | 'white' | 'gray'
    background TEXT NOT NULL DEFAULT 'black',
    FOREIGN KEY (user_id) REFERENCES user (id) ON DELETE CASCADE
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS reader_settings;
-- +goose StatementEnd

-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS library_source (
    library_id         TEXT NOT NULL,
    source             TEXT NOT NULL,
    enabled            INTEGER NOT NULL DEFAULT 0,
    cookies            TEXT,
    api_key            TEXT,
    magazine_blocklist TEXT NOT NULL DEFAULT '[]',
    misc_blocklist     TEXT NOT NULL DEFAULT '[]',
    PRIMARY KEY (library_id, source),
    FOREIGN KEY (library_id) REFERENCES library (id) ON DELETE CASCADE
);
-- +goose StatementEnd
-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS library_source;
-- +goose StatementEnd

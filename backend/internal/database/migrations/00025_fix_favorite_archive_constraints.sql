-- +goose Up
-- +goose StatementBegin
CREATE TABLE favorite_archive_new (
    archive_id TEXT NOT NULL,
    user_id TEXT NOT NULL,
    favorited_at DATETIME NOT NULL,
    PRIMARY KEY (archive_id, user_id),
    FOREIGN KEY (archive_id) REFERENCES archive (id) ON DELETE CASCADE,
    FOREIGN KEY (user_id) REFERENCES user (id) ON DELETE CASCADE
);
-- +goose StatementEnd
-- +goose StatementBegin
INSERT INTO favorite_archive_new (archive_id, user_id, favorited_at)
SELECT archive_id, user_id, favorited_at
FROM favorite_archive
WHERE archive_id IS NOT NULL AND user_id IS NOT NULL;
-- +goose StatementEnd
-- +goose StatementBegin
DROP TABLE favorite_archive;
-- +goose StatementEnd
-- +goose StatementBegin
ALTER TABLE favorite_archive_new RENAME TO favorite_archive;
-- +goose StatementEnd
-- +goose StatementBegin
CREATE INDEX IF NOT EXISTS idx_favorite_archive_user ON favorite_archive (user_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_favorite_archive_user;
-- +goose StatementEnd
-- +goose StatementBegin
CREATE TABLE favorite_archive_old (
    archive_id TEXT,
    user_id TEXT,
    favorited_at DATETIME NOT NULL,
    PRIMARY KEY (archive_id, user_id),
    FOREIGN KEY (archive_id) REFERENCES archive (id) ON DELETE CASCADE,
    FOREIGN KEY (user_id) REFERENCES user (id)
);
-- +goose StatementEnd
-- +goose StatementBegin
INSERT INTO favorite_archive_old (archive_id, user_id, favorited_at)
SELECT archive_id, user_id, favorited_at FROM favorite_archive;
-- +goose StatementEnd
-- +goose StatementBegin
DROP TABLE favorite_archive;
-- +goose StatementEnd
-- +goose StatementBegin
ALTER TABLE favorite_archive_old RENAME TO favorite_archive;
-- +goose StatementEnd

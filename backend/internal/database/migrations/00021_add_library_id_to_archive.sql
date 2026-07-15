-- +goose Up
-- +goose StatementBegin
ALTER TABLE archive ADD COLUMN library_id TEXT NOT NULL REFERENCES library (id) ON DELETE CASCADE;
-- +goose StatementEnd
-- +goose StatementBegin
CREATE INDEX IF NOT EXISTS idx_archive_library_id ON archive (library_id);
-- +goose StatementEnd
-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_archive_library_id;
-- +goose StatementEnd
-- +goose StatementBegin
ALTER TABLE archive DROP COLUMN library_id;
-- +goose StatementEnd

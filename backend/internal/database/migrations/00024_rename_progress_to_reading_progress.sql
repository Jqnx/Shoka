-- +goose Up
-- +goose StatementBegin
ALTER TABLE progress RENAME TO reading_progress;
-- +goose StatementEnd
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_progress_user_archive;
-- +goose StatementEnd
-- +goose StatementBegin
CREATE INDEX IF NOT EXISTS idx_reading_progress_user_archive ON reading_progress (user_id, archive_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE reading_progress RENAME TO progress;
-- +goose StatementEnd
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_reading_progress_user_archive;
-- +goose StatementEnd
-- +goose StatementBegin
CREATE INDEX IF NOT EXISTS idx_progress_user_archive ON progress (user_id, archive_id);
-- +goose StatementEnd

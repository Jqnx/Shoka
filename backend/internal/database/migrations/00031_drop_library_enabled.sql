-- +goose Up
-- +goose StatementBegin
-- The enable/disable toggle was removed: a library is either present (and
-- therefore watched/scanned) or deleted. Nothing read this column except
-- the watcher/scan guards that went with the feature.
ALTER TABLE library DROP COLUMN enabled;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE library ADD COLUMN enabled INTEGER NOT NULL DEFAULT 1;
-- +goose StatementEnd

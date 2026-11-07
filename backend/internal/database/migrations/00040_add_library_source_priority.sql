-- +goose Up
-- +goose StatementBegin
ALTER TABLE library_source ADD COLUMN priority INTEGER;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE library_source DROP COLUMN priority;
-- +goose StatementEnd

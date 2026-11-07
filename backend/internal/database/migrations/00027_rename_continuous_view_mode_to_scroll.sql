-- +goose Up
-- +goose StatementBegin
UPDATE reader_settings SET view_mode = 'scroll' WHERE view_mode = 'continuous';
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
UPDATE reader_settings SET view_mode = 'continuous' WHERE view_mode = 'scroll';
-- +goose StatementEnd

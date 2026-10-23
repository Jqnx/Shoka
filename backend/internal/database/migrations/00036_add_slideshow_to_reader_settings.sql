-- +goose Up
-- +goose StatementBegin
ALTER TABLE reader_settings ADD COLUMN slideshow_interval_seconds INTEGER NOT NULL DEFAULT 5;
-- +goose StatementEnd
-- +goose StatementBegin
ALTER TABLE reader_settings ADD COLUMN slideshow_loop BOOLEAN NOT NULL DEFAULT 1;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE reader_settings DROP COLUMN slideshow_loop;
-- +goose StatementEnd
-- +goose StatementBegin
ALTER TABLE reader_settings DROP COLUMN slideshow_interval_seconds;
-- +goose StatementEnd

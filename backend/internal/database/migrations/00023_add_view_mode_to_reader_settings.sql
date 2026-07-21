-- +goose Up
-- +goose StatementBegin
-- Orthogonal to reading_direction: view_mode picks paged vs. continuous
-- scroll, while reading_direction (ltr/rtl) only matters for tap/keyboard
-- navigation semantics within paged mode. Previously reading_direction had a
-- third 'vertical' value meant to double as continuous scroll - split out
-- into its own column instead so the two axes can vary independently.
ALTER TABLE reader_settings ADD COLUMN view_mode TEXT NOT NULL DEFAULT 'paged';
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE reader_settings DROP COLUMN view_mode;
-- +goose StatementEnd

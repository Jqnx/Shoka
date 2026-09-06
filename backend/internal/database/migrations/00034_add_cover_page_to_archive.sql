-- +goose Up
-- +goose StatementBegin
-- Which page (0-based) is used as the archive's cover thumbnail. Defaults
-- to 0 (the first page) - the historical, only behaviour before this column
-- existed. Set by the user via PUT /api/archives/{id}/cover; never touched
-- by the metadata pipeline (cover choice is presentation state, not
-- metadata that a source fetch should be able to overwrite).
ALTER TABLE archive ADD COLUMN cover_page INTEGER NOT NULL DEFAULT 0;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE archive DROP COLUMN cover_page;
-- +goose StatementEnd

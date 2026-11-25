-- +goose Up
-- +goose StatementBegin
-- Which image (0-based, ordered like GET /api/audio/{id}.images) is rendered
-- as the work's cover. Defaults to 0, the historical "first image" behaviour.
-- Set by the user via PUT /api/audio/{id}/cover; never touched by the metadata
-- editor (cover choice is presentation state, not metadata).
ALTER TABLE audio ADD COLUMN cover_index INTEGER NOT NULL DEFAULT 0;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE audio DROP COLUMN cover_index;
-- +goose StatementEnd

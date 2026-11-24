-- +goose Up
-- +goose StatementBegin
ALTER TABLE audio_played ADD COLUMN completed INTEGER NOT NULL DEFAULT 0;
-- +goose StatementEnd

-- Every row that predates automatic tracking came from a manual "mark played",
-- which means the user had finished the work.
-- +goose StatementBegin
UPDATE audio_played SET completed = 1;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE audio_played DROP COLUMN completed;
-- +goose StatementEnd

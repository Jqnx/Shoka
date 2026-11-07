-- +goose Up
-- +goose StatementBegin
-- Audio works can ship several versions as sibling folders (for example with
-- and without sound effects). This names the top-level folder an audio
-- library's works open on; empty means "no preference".
ALTER TABLE library ADD COLUMN default_audio_folder TEXT NOT NULL DEFAULT '';
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE library DROP COLUMN default_audio_folder;
-- +goose StatementEnd

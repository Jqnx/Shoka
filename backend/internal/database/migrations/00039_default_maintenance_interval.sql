-- +goose Up
-- +goose StatementBegin
UPDATE maintenance_task
SET interval_minutes = 720,
    updated_at = datetime('now');
-- +goose StatementEnd

-- +goose Down
UPDATE maintenance_task
SET interval_minutes = 0,
    updated_at = datetime('now');

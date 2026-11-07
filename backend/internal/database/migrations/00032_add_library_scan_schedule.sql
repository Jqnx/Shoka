-- +goose Up
-- +goose StatementBegin
-- Per-library scan scheduling. The fsnotify watcher misses changes on
-- network shares and bind mounts (no inotify propagation), so the timer acts
-- as a safety net and watch_enabled makes the watcher itself optional.
-- Defaults preserve existing behaviour: watcher on, timer off.
ALTER TABLE library ADD COLUMN scan_interval_minutes INTEGER NOT NULL DEFAULT 0;
ALTER TABLE library ADD COLUMN watch_enabled INTEGER NOT NULL DEFAULT 1;

-- NULL = never scanned, so due immediately. Distinct from updated_at, which
-- tracks config edits.
ALTER TABLE library ADD COLUMN last_scanned_at DATETIME;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE library DROP COLUMN last_scanned_at;
ALTER TABLE library DROP COLUMN watch_enabled;
ALTER TABLE library DROP COLUMN scan_interval_minutes;
-- +goose StatementEnd

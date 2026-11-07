-- +goose Up
-- +goose StatementBegin
CREATE TABLE maintenance_task (
    task            TEXT PRIMARY KEY NOT NULL,
    interval_minutes INTEGER NOT NULL DEFAULT 0,
    last_run_at     DATETIME,
    created_at      DATETIME NOT NULL DEFAULT (datetime('now')),
    updated_at      DATETIME NOT NULL DEFAULT (datetime('now'))
);
INSERT INTO maintenance_task (task) VALUES ('cover'), ('phash');
-- +goose StatementEnd

-- +goose Down
DROP TABLE maintenance_task;

-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS job (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    type        TEXT    NOT NULL,
    payload     TEXT    NOT NULL DEFAULT '{}',
    status      TEXT    NOT NULL DEFAULT 'pending'
                        CHECK(status IN ('pending', 'running', 'done', 'failed')),
    attempts    INTEGER NOT NULL DEFAULT 0,
    max_attempts INTEGER NOT NULL DEFAULT 3,
    error       TEXT,
    created_at  DATETIME NOT NULL DEFAULT (datetime('now')),
    updated_at  DATETIME NOT NULL DEFAULT (datetime('now')),
    run_after   DATETIME NOT NULL DEFAULT (datetime('now'))
);

CREATE INDEX IF NOT EXISTS idx_job_status_run_after
ON job (status, run_after);
-- +goose StatementEnd
-- +goose Down
-- +goose StatementBegin
drop table if exists job;
-- +goose StatementEnd

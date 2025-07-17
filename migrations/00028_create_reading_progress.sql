-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS reading_progress (
  archive_id bigint,
  user_id bigint,
  page bigint NOT NULL,
  state text NOT NULL,
  last_read timestamptz NOT NULL,
  PRIMARY KEY (archive_id, user_id),
  FOREIGN KEY (archive_id) REFERENCES archives(id),
  FOREIGN KEY (user_id) REFERENCES users(id)
);
-- +goose StatementEnd
-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS reading_progress;
-- +goose StatementEnd



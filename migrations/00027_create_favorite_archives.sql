-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS favorite_archives (
  archive_id bigint,
  user_id bigint,
  favorited_at timestamptz NOT NULL,
  PRIMARY KEY (archive_id, user_id),
  FOREIGN KEY (archive_id) REFERENCES archives(id),
  FOREIGN KEY (user_id) REFERENCES users(id)
);
-- +goose StatementEnd
-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS favorite_archives;
-- +goose StatementEnd



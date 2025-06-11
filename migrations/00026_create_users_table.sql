-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS users (
  id bigserial PRIMARY KEY,
  name text NOT NULL UNIQUE,
  password text NOT NULL,
  session text,
  session_expiry timestamptz,
  created_at timestamptz NOT NULL,
  updated_at timestamptz NOT NULL
);
-- +goose StatementEnd
-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS users;
-- +goose StatementEnd



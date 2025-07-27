-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS sessions (
  session_id uuid primary key,
  user_id uuid not null references users(id) on delete cascade,
  token text not null unique,
  expires_at timestamptz not null,
  created_at timestamptz not null,
  ip_address inet,
  user_agent text
);
-- +goose StatementEnd
-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS sessions;
-- +goose StatementEnd



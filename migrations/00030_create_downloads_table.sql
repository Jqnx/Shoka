-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS downloads (
  id uuid PRIMARY KEY,
  url text NOT NULL,
  filename text NOT NULL,
  status text NOT NULL,
  progress integer DEFAULT 0,
  error text,
  created_at timestamptz NOT NULL,
  updated_at timestamptz NOT NULL,
  speed bigint,
  total_size bigint,
  downloaded bigint,
  started_at timestamptz,
  can_resume boolean,
  resume_supported boolean
);

CREATE INDEX IF NOT EXISTS idx_downloads_status ON downloads (status);
-- +goose StatementEnd
-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS downloads;
-- +goose StatementEnd



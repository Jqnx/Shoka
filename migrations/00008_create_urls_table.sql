-- +goose Up
CREATE TABLE IF NOT EXISTS urls (
    id bigserial PRIMARY KEY,
    url text NOT NULL,
    archive_id bigint NOT NULL,
    FOREIGN KEY (archive_id) REFERENCES archives (id)
);

-- +goose Down
DROP TABLE IF EXISTS urls;

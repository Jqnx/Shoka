-- +goose Up
CREATE TABLE IF NOT EXISTS urls (
    id bigserial PRIMARY KEY,
    url text NOT NULL,
    archive_id bigint NOT NULL,
    FOREIGN KEY (archive_id) REFERENCES archives (id)
);

create index idx_url_archive_id on urls(archive_id);

-- +goose Down
DROP TABLE IF EXISTS urls;


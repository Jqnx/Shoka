-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS archive_circle (
    archive_id TEXT,
    circle_id INTEGER,
    PRIMARY KEY (archive_id, circle_id),
    FOREIGN KEY (archive_id) REFERENCES archive (id) ON DELETE CASCADE,
    FOREIGN KEY (circle_id) REFERENCES circle (id) ON DELETE CASCADE
);
-- +goose StatementEnd
-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS archive_circle;
-- +goose StatementEnd

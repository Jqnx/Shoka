-- +goose Up
-- +goose StatementBegin
CREATE TABLE library_access (
    user_id TEXT NOT NULL REFERENCES user (id) ON DELETE CASCADE,
    library_id TEXT NOT NULL REFERENCES library (id) ON DELETE CASCADE,
    PRIMARY KEY (user_id, library_id)
);
CREATE INDEX idx_library_access_library ON library_access (library_id);
-- +goose StatementEnd

-- +goose Down
DROP TABLE library_access;

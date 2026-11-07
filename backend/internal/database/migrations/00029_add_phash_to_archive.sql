-- +goose Up
-- +goose StatementBegin
-- Stores a 64-bit perceptual hash (DCT-based pHash of the cover/page 0) as a
-- bit-cast signed INTEGER - SQLite has no native uint64, so the Go side
-- reinterprets bits both ways (uint64(v) / int64(v)), never treats it as a
-- real signed magnitude. NULL until the cover job computes it.
ALTER TABLE archive ADD COLUMN phash INTEGER;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE archive DROP COLUMN phash;
-- +goose StatementEnd

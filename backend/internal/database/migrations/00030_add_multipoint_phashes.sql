-- +goose Up
-- +goose StatementBegin
-- Cover-only duplicate detection was both over- and under-sensitive: two
-- unrelated works with a similar cover matched, while a re-release with a
-- new cover but identical interior didn't. Sample four points through the
-- archive instead - p0 (the cover) plus 25/50/75% - so a match has to hold
-- across the body of the work, not just its first page.
--
-- The existing `phash` column already holds the p0 hash, so it's renamed
-- rather than dropped and recomputed.
ALTER TABLE archive RENAME COLUMN phash TO phash_p0;
-- +goose StatementEnd
-- +goose StatementBegin
ALTER TABLE archive ADD COLUMN phash_p25 INTEGER;
-- +goose StatementEnd
-- +goose StatementBegin
ALTER TABLE archive ADD COLUMN phash_p50 INTEGER;
-- +goose StatementEnd
-- +goose StatementBegin
ALTER TABLE archive ADD COLUMN phash_p75 INTEGER;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE archive DROP COLUMN phash_p75;
-- +goose StatementEnd
-- +goose StatementBegin
ALTER TABLE archive DROP COLUMN phash_p50;
-- +goose StatementEnd
-- +goose StatementBegin
ALTER TABLE archive DROP COLUMN phash_p25;
-- +goose StatementEnd
-- +goose StatementBegin
ALTER TABLE archive RENAME COLUMN phash_p0 TO phash;
-- +goose StatementEnd

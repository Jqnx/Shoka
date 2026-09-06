-- name: CreateArchive :one
insert into archive (
  id,
  library_id,
  title,
  file_path,
  file_size,
  mod_time,
  page_count
)
values (?, ?, ?, ?, ?, ?, ?)
returning *;


-- name: GetArchiveByID :one
select *
from archive
where id = ?
;

-- name: GetArchiveByFilePath :one
select *
from archive
where file_path = ?
;

-- name: GetAllArchiveFilePaths :many
-- returns file paths for every archive across all libraries; used for
-- cross-library maintenance actions (e.g. regenerating all covers).
-- cover_page rides along so the maintenance "regenerate covers" actions can
-- re-run the cover job against the user's chosen page instead of resetting
-- every archive back to page 0.
select id, file_path, file_size, mod_time, cover_page
from archive
;

-- name: GetArchiveFilePathsByLibrary :many
-- scoped to a single library so the scanner never mistakes another
-- library's archives for files that were removed from disk.
select id, file_path, file_size, mod_time, cover_page
from archive
where library_id = ?
;

-- name: GetAllArchives :many
select archive.*, reading_progress.page, reading_progress.last_read, reading_progress.completed
from archive
left join
    reading_progress
    on archive.id = reading_progress.archive_id
    and reading_progress.user_id = sqlc.arg('uid')
where archive.library_id = sqlc.arg('library_id')
order by archive.id
;

-- name: GetArchiveList :many
select archive.*, reading_progress.page, reading_progress.last_read, reading_progress.completed
from archive
left join
    reading_progress on archive.id = reading_progress.archive_id and reading_progress.user_id = sqlc.arg('uid')
where archive.library_id = sqlc.arg('library_id')
limit sqlc.arg('limit')
offset sqlc.arg('offset')
;

-- name: GetArchiveShuffle :one
select id
from archive
where library_id = ?
limit ?
offset ?
;

-- name: ArchiveExists :execresult
select title
from archive
where title = ?
;

-- name: FilePathExists :one
select exists (select 1 from archive where file_path = ?) = 1
;

-- name: ArchiveIDExists :execresult
select id
from archive
where id = ?
;

-- name: GetFilePathByID :one
select file_path
from archive
where id = ?
;

-- name: CountArchives :one
select count(*)
from archive
;

-- name: UpdateArchive :exec
update archive
set title = coalesce(sqlc.narg('title'), title),
    summary = coalesce(sqlc.narg('summary'), summary),
    language = coalesce(sqlc.narg('language'), language),
    category = coalesce(sqlc.narg('category'), category),
    page_count = coalesce(sqlc.narg('page_count'), page_count),
    release_date = coalesce(sqlc.narg('release_date'), release_date),
    updated_at = datetime('now')
where id = sqlc.arg('id')
;

-- name: UpdateArchiveMeta :exec
update archive
set
    file_size = ?,
    mod_time = ?,
    updated_at = datetime('now')
where id = ?;

-- name: UpdateFilePath :exec
update archive
set file_path = ?,
    file_size = ?,
    mod_time = ?,
    updated_at = datetime('now')
where id = ?;

-- name: UpdateArchiveCoverPage :exec
-- cover_page is presentation state, not metadata - deliberately its own
-- query rather than folded into UpdateArchive, so a metadata source fetch
-- (which goes through UpdateArchive/ApplyMetadata) can never clobber it.
-- updated_at is left untouched on purpose: picking a cover page is not a
-- content change and must not resurface the archive in "recently updated".
update archive
set cover_page = ?
where id = ?;

-- name: DeleteArchive :exec
delete from archive
where id = ?
;

-- name: DeleteAllArchive :exec
delete from archive
;

-- name: DeleteArchiveByFilePath :exec
delete from archive
where file_path = ?
;

-- name: UpdateArchivePHashes :exec
-- All four sample points are written together by the phash job, so there's
-- no partial-update path to worry about.
update archive
set phash_p0 = ?,
    phash_p25 = ?,
    phash_p50 = ?,
    phash_p75 = ?
where id = ?
;

-- name: GetAllArchivePHashes :many
-- Deliberately NOT scoped to a library - duplicate detection is a
-- cross-library maintenance action (the same doujin can end up scanned
-- into two different library folders). Rows with no hashes at all are
-- excluded; a row with only some points hashed is still useful (the
-- comparison just has fewer points to work with).
select id, title, library_id, cover_page, phash_p0, phash_p25, phash_p50, phash_p75
from archive
where phash_p0 is not null
   or phash_p25 is not null
   or phash_p50 is not null
   or phash_p75 is not null
;

-- name: GetArchivesByIDs :many
-- Backs the bulk endpoints: resolves which of a caller-supplied set of ids
-- actually exist (so missing ones can be reported per-id rather than
-- failing the whole batch) and carries page_count for read-marking.
select id, page_count
from archive
where id in (sqlc.slice('ids'))
;

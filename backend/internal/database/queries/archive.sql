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
select id, file_path, file_size, mod_time
from archive
;

-- name: GetArchiveFilePathsByLibrary :many
-- scoped to a single library so the scanner never mistakes another
-- library's archives for files that were removed from disk.
select id, file_path, file_size, mod_time
from archive
where library_id = ?
;

-- name: GetAllArchives :many
select archive.*, progress.page, progress.last_read, progress.completed
from archive
left join
    progress
    on archive.id = progress.archive_id
    and progress.user
    and progress.user_id = sqlc.arg('uid')
where archive.library_id = sqlc.arg('library_id')
order by archive.id
;

-- name: GetRecentlyReadArchives :many
select archive.*, progress.page, progress.last_read, progress.completed
from archive
left join
    progress on archive.id = progress.archive_id and progress.user_id = sqlc.arg('uid')
where progress.last_read is not null
    and archive.library_id = sqlc.arg('library_id')
order by progress.last_read
;

-- name: GetArchiveList :many
select archive.*, progress.page, progress.last_read, progress.completed
from archive
left join
    progress on archive.id = progress.archive_id and progress.user_id = sqlc.arg('uid')
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

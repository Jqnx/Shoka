-- name: CreateArchive :one
insert into archive (
  id,
  title,
  file_path,
  file_size,
  mod_time
)
values (?, ?, ?, ?, ?)
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
select id, file_path, file_size, mod_time
from archive
;

-- name: GetAllArchives :many
select
    archive.*,
    reading_progress.page,
    reading_progress.last_read,
    reading_progress.status
from archive
left join
    reading_progress
    on archive.id = reading_progress.archive_id
    and reading_progress.user
    and reading_progress.user_id = sqlc.arg('uid')
order by archive.id
;

-- name: GetRecentlyReadArchives :many
select
    archive.*,
    reading_progress.page,
    reading_progress.last_read,
    reading_progress.status
from archive
left join
    reading_progress
    on archive.id = reading_progress.archive_id
    and reading_progress.user_id = sqlc.arg('uid')
where reading_progress.last_read is not null
order by reading_progress.last_read
;

-- name: GetArchiveList :many
select
    archive.*,
    reading_progress.page,
    reading_progress.last_read,
    reading_progress.status
from archive
left join
    reading_progress
    on archive.id = reading_progress.archive_id
    and reading_progress.user_id = sqlc.arg('uid')
limit ?
offset ?
;

-- name: GetArchiveShuffle :one
select id
from archive
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

-- name: UpdateArchive :one
update archive
set title = coalesce(sqlc.narg('title'), title),
    summary = coalesce(sqlc.narg('summary'), summary),
    language = coalesce(slqc.narg('language'), language),
    category = coalesce(sqlc.narg('category'), category),
    page_count = coalesce(sqlc.narg('page_count'), page_count),
    updated_at = datetime('now'),
    release_date = coalesce(sqlc.narg('release_date'), release_date)
where id = ?
returning *
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
    updated_at = ?
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

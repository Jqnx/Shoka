-- name: CreateArchive :one
insert into archives (
    title,
    summary,
    language,
    category,
    page_count,
    file_path,
    archive_id,
    hash,
    thumbs_path,
    cover_path,
-- pages_path,
    type,
    created_at,
    updated_at,
    release_date
    )
values ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
returning *
;

-- name: GetLastArchiveID :one
select archive_id
from archives
order by archive_id desc
limit 1
;

-- name: GetArchiveByID :one
select *
from archives
where archive_id = $1
;

-- name: GetArchiveByFilePath :one
select *
from archives
where file_path = $1
;

-- name: GetAllArchives :many
select *
from archives
order by archive_id
;

-- name: GetArchiveList :many
select *
from archives
limit $1
offset $2
;

-- name: ArchiveExists :execresult
select title
from archives
where title = $1
;

-- name: FilePathExists :execresult
select file_path
from archives
where file_path = $1
;

-- name: ArchiveIDExists :execresult
select archive_id
from archives
where archive_id = $1
;

-- name: ThumbsPathExistsForFilePath :execresult
select thumbs_path
from archives
where file_path = $1
;

-- name: CountArchives :one
select count(*)
from archives
;

-- name: UpdateArchive :one
update archives
set title = coalesce(sqlc.narg('title'), title),
    summary = coalesce(sqlc.narg('summary'), summary),
    language = coalesce(sqlc.narg('language'), language),
    category = coalesce(sqlc.narg('category'), category),
    updated_at = coalesce($1, updated_at),
    release_date = coalesce(sqlc.narg('release_date'), release_date)
where archive_id = $2
returning *
;

-- name: UpdateThumbPath :exec
update archives
set thumbs_path = $1
where archive_id = $2
;

-- name: UpdateCoverPath :exec
update archives
set cover_path = $1
where archive_id = $2
;

-- name: DeleteArchive :exec
delete from archives
where archive_id = $1
;


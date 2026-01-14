-- name: CreateArchive :one
insert into archive (
    id,
    title,
    summary,
    language,
    category,
    page_count,
    file_path,
    file_hash,
    thumb_path,
    created_at,
    updated_at,
    release_date
    )
values ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
returning
    id,
    title,
    summary,
    language,
    category,
    page_count,
    file_path,
    file_hash,
    thumb_path,
    created_at,
    updated_at,
    release_date
;

-- name: GetArchiveByID :one
select
    id,
    title,
    summary,
    language,
    category,
    page_count,
    file_path,
    file_hash,
    thumb_path,
    created_at,
    updated_at,
    release_date
from archive
where id = $1
;

-- name: GetArchiveByFilePath :one
select
    id,
    title,
    summary,
    language,
    category,
    page_count,
    file_path,
    file_hash,
    thumb_path,
    created_at,
    updated_at,
    release_date
from archive
where file_path = $1
;

-- name: GetArchiveByFileHash :one
select
    id,
    title,
    summary,
    language,
    category,
    page_count,
    file_path,
    file_hash,
    thumb_path,
    created_at,
    updated_at,
    release_date
from archive
where file_hash = $1
;

-- name: GetAllFilePaths :many
select id, file_path
from archive
;

-- name: GetAllArchives :many
select
    archive.id,
    archive.title,
    archive.summary,
    archive.language,
    archive.category,
    archive.page_count,
    archive.file_path,
    archive.file_hash,
    archive.thumb_path,
    archive.created_at,
    archive.updated_at,
    archive.release_date,
    reading_progress.page,
    reading_progress.last_read,
    reading_progress.status
from archive
left join
    reading_progress
    on archive.id = reading_progress.archive_id
    and reading_progress.user_id = sqlc.arg('uid')
order by archive.id
;

-- name: GetRecentlyReadArchives :many
select
    archive.id,
    archive.title,
    archive.summary,
    archive.language,
    archive.category,
    archive.page_count,
    archive.file_path,
    archive.file_hash,
    archive.thumb_path,
    archive.created_at,
    archive.updated_at,
    archive.release_date,
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
    archive.id,
    archive.title,
    archive.summary,
    archive.language,
    archive.category,
    archive.page_count,
    archive.file_path,
    archive.file_hash,
    archive.thumb_path,
    archive.created_at,
    archive.updated_at,
    archive.release_date,
    reading_progress.page,
    reading_progress.last_read,
    reading_progress.status
from archive
left join
    reading_progress
    on archive.id = reading_progress.archive_id
    and reading_progress.user_id = sqlc.arg('uid')
limit $1
offset $2
;

-- name: GetArchiveShuffle :one
select id
from archive
limit $1
offset $2
;

-- name: SearchArchive :many
select
    id,
    title,
    summary,
    language,
    category,
    page_count,
    file_path,
    file_hash,
    thumb_path,
    created_at,
    updated_at,
    release_date,
    ts_rank_cd(search_vector, query) as rank
from archive, websearch_to_tsquery('english', $1) query
where search_vector @@ query
order by rank desc
;

-- name: CountSearchArchive :many
select count(*)
from archive, websearch_to_tsquery('english', $1) query
where search_vector @@ query
;

-- name: SearchArchiveList :many
select
    id,
    title,
    summary,
    language,
    category,
    page_count,
    file_path,
    file_hash,
    thumb_path,
    created_at,
    updated_at,
    release_date,
    ts_rank_cd(search_vector, query) as rank
from archive, websearch_to_tsquery('english', $1) query
where search_vector @@ query
order by rank desc
limit $2
offset $3
;

-- name: ArchiveExists :execresult
select title
from archive
where title = $1
;

-- name: FilePathExists :execresult
select file_path
from archive
where file_path = $1
;

-- name: ArchiveIDExists :execresult
select id
from archive
where id = $1
;

-- name: ThumbsPathExistsForFilePath :execresult
select thumb_path
from archive
where file_path = $1
;

-- name: GetThumbPathByID :one
select thumb_path
from archive
where id = $1
;

-- name: GetFilePathByID :one
select file_path
from archive
where id = $1
;

-- name: CountArchives :one
select count(*)
from archive
;

-- name: UpdateArchive :one
update archive
set title = coalesce(sqlc.narg('title'), title),
    summary = coalesce(sqlc.narg('summary'), summary),
    language = coalesce(sqlc.narg('language'), language),
    category = coalesce(sqlc.narg('category'), category),
    updated_at = coalesce($1, updated_at),
    release_date = coalesce(sqlc.narg('release_date'), release_date)
where id = $2
returning
    id,
    title,
    summary,
    language,
    category,
    page_count,
    file_path,
    file_hash,
    thumb_path,
    created_at,
    updated_at,
    release_date
;

-- name: UpdateThumbPath :exec
update archive
set thumb_path = $1,
    updated_at = $2
where id = $3
;

-- name: UpdateFilePath :exec
update archive
set file_path = $1,
    updated_at = $2
where id = $3
;

-- name: UpdateFileHash :exec
update archive
set file_hash = $1,
    updated_at = $2
where id = $3
;

-- name: DeleteArchive :exec
delete from archive
where id = $1
;

-- name: DeleteAllArchive :exec
delete from archive
;

-- name: DeleteArchiveByFilePath :exec
delete from archive
where file_path = $1
;


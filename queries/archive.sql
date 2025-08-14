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
    type,
    created_at,
    updated_at,
    release_date
    )
values ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
returning
    id,
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
    type,
    created_at,
    updated_at,
    release_date
;

-- name: GetLastArchiveID :one
select archive_id
from archives
order by archive_id desc
limit 1
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
    archive_id,
    hash,
    thumbs_path,
    cover_path,
    type,
    created_at,
    updated_at,
    release_date
from archives
where archive_id = $1
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
    archive_id,
    hash,
    thumbs_path,
    cover_path,
    type,
    created_at,
    updated_at,
    release_date
from archives
where file_path = $1
;

-- name: GetAllArchives :many
select
    archives.id,
    archives.title,
    archives.summary,
    archives.language,
    archives.category,
    archives.page_count,
    archives.file_path,
    archives.archive_id,
    archives.hash,
    archives.thumbs_path,
    archives.cover_path,
    archives.type,
    archives.created_at,
    archives.updated_at,
    archives.release_date,
    reading_progress.page,
    reading_progress.last_read,
    reading_progress.state
from archives
left join
    reading_progress
    on archives.id = reading_progress.archive_id
    and reading_progress.user_id = sqlc.arg('uid')
order by archives.archive_id
;

-- name: GetRecentlyReadArchives :many
select
    archives.id,
    archives.title,
    archives.summary,
    archives.language,
    archives.category,
    archives.page_count,
    archives.file_path,
    archives.archive_id,
    archives.hash,
    archives.thumbs_path,
    archives.cover_path,
    archives.type,
    archives.created_at,
    archives.updated_at,
    archives.release_date,
    reading_progress.page,
    reading_progress.last_read,
    reading_progress.state
from archives
left join
    reading_progress
    on archives.id = reading_progress.archive_id
    and reading_progress.user_id = sqlc.arg('uid')
where reading_progress.last_read is not null
order by reading_progress.last_read
;

-- name: GetArchiveList :many
select
    archives.id,
    archives.title,
    archives.summary,
    archives.language,
    archives.category,
    archives.page_count,
    archives.file_path,
    archives.archive_id,
    archives.hash,
    archives.thumbs_path,
    archives.cover_path,
    archives.type,
    archives.created_at,
    archives.updated_at,
    archives.release_date,
    reading_progress.page,
    reading_progress.last_read,
    reading_progress.state
from archives
left join
    reading_progress
    on archives.id = reading_progress.archive_id
    and reading_progress.user_id = sqlc.arg('uid')
limit $1
offset $2
;

-- name: GetArchiveShuffle :one
select archive_id
from archives
limit $1
offset $2
;

-- name: SearchArchives :many
select
    id,
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
    type,
    created_at,
    updated_at,
    release_date,
    ts_rank_cd(search_vector, query) as rank
from archives, websearch_to_tsquery('english', $1) query
where search_vector @@ query
order by rank desc
;

-- name: CountSearchArchives :many
select count(*)
from archives, websearch_to_tsquery('english', $1) query
where search_vector @@ query
;

-- name: SearchArchivesList :many
select
    id,
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
    type,
    created_at,
    updated_at,
    release_date,
    ts_rank_cd(search_vector, query) as rank
from archives, websearch_to_tsquery('english', $1) query
where search_vector @@ query
order by rank desc
limit $2
offset $3
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
returning
    id,
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
    type,
    created_at,
    updated_at,
    release_date
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


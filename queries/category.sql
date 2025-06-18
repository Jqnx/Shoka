-- name: GetAllCategory :many
select category
from archives
where category is not null
;

-- name: GetArchivesByCategory :many
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
    archives.release_date
from archives
where category = $1
;

-- name: GetArchiveIDsByCategory :many
select archives.archive_id
from archives
where category = $1
;

-- name: GetArchivesByCategoryList :many
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
    archives.release_date
from archives
where category = $1
limit $2
offset $3
;

-- name: TotalArchivesWithCategory :one
select count(archive_id)
from archives
where category = $1
;

-- name: CategoryExists :execresult
select category
from archives
where category = $1
;


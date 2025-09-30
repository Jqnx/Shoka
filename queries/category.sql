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
    archives.file_name,
    -- archives.archive_id,
    archives.hash,
    archives.thumbs_path,
    archives.cover_path,
    archives.type,
    archives.created_at,
    archives.updated_at,
    archives.release_date,
    reading_progress.page
from archives
left join
    reading_progress
    on archives.id = reading_progress.archive_id
    and reading_progress.user_id = sqlc.arg('uid')
where category = $1
;

-- name: GetArchiveIDsByCategory :many
select archives.id
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
    archives.file_name,
    -- archives.archive_id,
    archives.hash,
    archives.thumbs_path,
    archives.cover_path,
    archives.type,
    archives.created_at,
    archives.updated_at,
    archives.release_date,
    reading_progress.page
from archives
left join
    reading_progress
    on archives.id = reading_progress.archive_id
    and reading_progress.user_id = sqlc.arg('uid')
where category = $1
limit $2
offset $3
;

-- name: TotalArchivesWithCategory :one
select count(id)
from archives
where category = $1
;

-- name: CategoryExists :execresult
select category
from archives
where category = $1
;


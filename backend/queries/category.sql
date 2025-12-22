-- name: GetAllCategory :many
select category
from archive
where category is not null
;

-- name: GetArchiveByCategory :many
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
    reading_progress.page
from archive
left join
    reading_progress
    on archive.id = reading_progress.archive_id
    and reading_progress.user_id = sqlc.arg('uid')
where category = $1
;

-- name: GetArchiveIDsByCategory :many
select archive.id
from archive
where category = $1
;

-- name: GetArchiveByCategoryList :many
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
    reading_progress.page
from archive
left join
    reading_progress
    on archive.id = reading_progress.archive_id
    and reading_progress.user_id = sqlc.arg('uid')
where category = $1
limit $2
offset $3
;

-- name: TotalArchiveWithCategory :one
select count(id)
from archive
where category = $1
;

-- name: CategoryExists :execresult
select category
from archive
where category = $1
;


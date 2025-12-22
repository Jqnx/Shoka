-- name: GetAllLanguage :many
select language
from archive
where language is not null
;

-- name: GetArchiveByLanguage :many
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
where language = $1
;

-- name: GetArchiveIDsByLanguage :many
select archive.id
from archive
where language = $1
;

-- name: GetArchiveByLanguageList :many
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
where language = $1
limit $2
offset $3
;

-- name: TotalArchiveWithLanguage :one
select count(id)
from archive
where language = $1
;

-- name: LanguageExists :execresult
select language
from archive
where language = $1
;


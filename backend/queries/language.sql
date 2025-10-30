-- name: GetAllLanguage :many
select language
from archives
where language is not null
;

-- name: GetArchivesByLanguage :many
select
    archives.id,
    archives.title,
    archives.summary,
    archives.language,
    archives.category,
    archives.page_count,
    archives.file_path,
    archives.file_name,
    archives.hash,
    archives.thumbs_path,
    archives.cover_path,
    archives.cover_img,
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
where language = $1
;

-- name: GetArchiveIDsByLanguage :many
select archives.id
from archives
where language = $1
;

-- name: GetArchivesByLanguageList :many
select
    archives.id,
    archives.title,
    archives.summary,
    archives.language,
    archives.category,
    archives.page_count,
    archives.file_path,
    archives.file_name,
    archives.hash,
    archives.thumbs_path,
    archives.cover_path,
    archives.cover_img,
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
where language = $1
limit $2
offset $3
;

-- name: TotalArchivesWithLanguage :one
select count(id)
from archives
where language = $1
;

-- name: LanguageExists :execresult
select language
from archives
where language = $1
;


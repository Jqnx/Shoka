-- name: GetAllLanguage :many
select language
from archives
where language is not null
order by language asc
;

-- name: GetArchivesByLanguage :many
select *
from archives
where language = $1
;

-- name: GetArchivesByLanguageList :many
select *
from archives
where language = $1
limit $2
offset $3
;

-- name: TotalArchivesWithLanguage :one
select count(archive_id)
from archives
where language = $1
;

-- name: LanguageExists :execresult
select *
from archives
where language = $1
;


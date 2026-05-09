-- name: GetAllLanguage :many
select distinct language
from archive
where language is not null
;

-- name: GetArchiveByLanguage :many
select archive.*, progress.page
from archive
left join
    progress on archive.id = progress.archive_id and progress.user_id = sqlc.arg('uid')
where language = ?
;

-- name: GetArchiveIDsByLanguage :many
select archive.id
from archive
where language = ?
;

-- name: GetArchiveIDsByLanguages :many
select archive.id
from archive
where language in (sqlc.slice('languages'))
group by archive.id
having count(distinct archive.id) = sqlc.arg('amount')
;

-- name: GetArchiveByLanguageList :many
select archive.*, progress.page
from archive
left join
    progress on archive.id = progress.archive_id and progress.user_id = sqlc.arg('uid')
where language = ?
order by
    case when sqlc.arg('order_by') = 'title_asc' then archive.title end asc,
    case when sqlc.arg('order_by') = 'title_desc' then archive.title end desc,
    case when sqlc.arg('order_by') = 'page_count_asc' then archive.page_count end asc,
    case when sqlc.arg('order_by') = 'page_count_desc' then archive.page_count end desc,
    case when sqlc.arg('order_by') = 'created_at_asc' then archive.created_at end asc,
    case when sqlc.arg('order_by') = 'created_at_desc' then archive.created_at end desc,
    case when sqlc.arg('order_by') = 'updated_at_asc' then archive.updated_at end asc,
    case when sqlc.arg('order_by') = 'updated_at_desc' then archive.updated_at end desc,
    case
        when sqlc.arg('order_by') = 'release_date_asc' then archive.release_date
    end asc,
    case
        when sqlc.arg('order_by') = 'release_date_desc' then archive.release_date
    end desc,
    case when sqlc.arg('order_by') = 'last_read_asc' then progress.last_read end asc,
    case when sqlc.arg('order_by') = 'last_read_desc' then progress.last_read end desc
limit ?
offset ?
;

-- name: TotalArchiveWithLanguage :one
select count(id)
from archive
where language = ?
;

-- name: LanguageExists :execresult
select language
from archive
where language = ?
;

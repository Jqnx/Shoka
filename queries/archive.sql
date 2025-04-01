-- name: CreateArchive :one
insert into archives (
    title,
    summary,
    lang,
    category,
    page_count,
    file_path,
    a_id,
    created_at,
    updated_at
    )
values ($1, $2, $3, $4, $5, $6, $7, $8, $9)
returning *
;

-- name: GetArchiveLastAID :one
select a_id
from archives
order by a_id desc
limit 1
;

-- name: GetArchiveByAID :one
select *
from archives
where a_id = $1
;

-- name: GetAllArchives :many
select *
from archives
order by a_id
;


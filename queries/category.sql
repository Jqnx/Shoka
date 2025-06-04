-- name: GetAllCategory :many
select category
from archives
where category is not null
;

-- name: GetArchivesByCategory :many
select *
from archives
where category = $1
;

-- name: GetArchivesByCategoryList :many
select *
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
select *
from archives
where category = $1
;


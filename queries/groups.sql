-- name: CreateGroup :one
insert into groups (name)
values ($1)
returning *
;

-- name: GetAllGroups :many
select *
from groups
order by id
;

-- name: GetGroup :one
select *
from groups
where name = $1
;

-- name: GroupExists :execresult
select *
from groups
where name = $1
;


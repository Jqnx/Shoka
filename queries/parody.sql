-- name: CreateParody :one
insert into parodies (parody)
values ($1)
returning *
;

-- name: AddParodyToArchive :exec
insert into archives_parodies (archive_id, parody_id)
values ($1, $2)
;

-- name: GetParody :one
select id, parody
from parodies
where parody = $1
;

-- name: GetAllParodies :many
select id, parody
from parodies
order by id
;

-- name: ParodyExists :execresult
select id, parody
from parodies
where parody = $1
;

-- name: RemoveParodyFromArchive :exec
delete from archives_parodies
where archive_id = $1
;


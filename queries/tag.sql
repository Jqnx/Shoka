-- name: CreateTag :one
insert into tags (tag)
values ( $1 )
returning *
;

-- name: AddTagToArchive :exec
insert into archives_tags (archive_id, tag_id)
values ($1, $2)
;

-- name: GetTag :one
select id, tag
from tags
where tag = $1
;

-- name: GetAllTags :many
select id, tag
from tags
order by id
;

-- name: TagExists :execresult
select id, tag
from tags
where tag = $1
;

-- name: RemoveTagFromArchive :exec
delete from archives_tags
where archive_id = $1
;


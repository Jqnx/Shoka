-- name: CreateTag :one
insert into tags (tag)
values ( $1 )
returning *
;

-- name: AddTagToArchive :exec
insert into archives_tags (archive_id, tag_id)
values ($1, $2)
;

-- name: GetAllTags :many
select *
from tags
order by id
;


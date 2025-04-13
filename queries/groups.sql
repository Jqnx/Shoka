-- name: CreateGroup :one
insert into groups (name, created_at, updated_at)
values ($1, $2, $3)
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

-- name: GetGroupArtists :many
select artists.name
from groups
join artists_groups on groups.id = artists_groups.group_id
join artists on artists_groups.artist_id = artists.id
where groups.name = $1
;

-- name: GroupExists :execresult
select *
from groups
where name = $1
;

-- name: UpdateGroup :one
update groups
set name = $1,
    updated_at = $2
where name = sqlc.arg(old_name)::text
returning *
;

-- name: RemoveArtistsFromGroup :exec
delete from artists_groups
where group_id = $1
;

-- name: DeleteGroup :exec
delete from groups
where name = $1
;


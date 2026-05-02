-- name: CreateCircle :one
insert into circle (name, count)
values (?, ?)
returning *
;

-- name: GetAllCircle :many
select *
from circle
order by id
;

-- name: GetCircle :one
select *
from circle
where name = ?
;

-- name: GetCircleList :many
select *
from circle
order by ?
limit ?
offset ?
;

-- name: GetArchiveCircle :many
select circle.*
from archive
join archive_circle on archive.id = archive_circle.archive_id
join circle on archive_circle.circle_id = circle.id
where archive.id = ?
;

-- name: GetArchiveCircleIDs :many
select circle.id
from archive
join archive_circle on archive.id = archive_circle.archive_id
join circle on archive_circle.circle_id = circle.id
where archive.id = ?
;

-- name: GetCircleArtists :many
select artist.*
from circle
join artist_circle on circle.id = artist_circle.circle_id
join artist on artist_circle.artist_id = artist.id
where circle.name = ?
;

-- name: CircleExists :execresult
select name
from circle
where name = ?
;

-- name: TotalCircle :one
select count(id)
from circle
;

-- name: UpdateCircle :one
update circle
set name = ?
where name = sqlc.arg('old_name')
returning *
;

-- name: RemoveArtistsFromCircle :exec
delete from artist_circle
where circle_id = ?
;

-- name: DeleteCircle :exec
delete from circle
where name = ?
;

-- name: BulkAddArchiveCircle :exec
insert or ignore into archive_circle (archive_id, circle_id)
select ?, value
from json_each(sqlc.arg('circle'))
;

-- name: BulkAddCircle :exec
insert or ignore into circle (name, count)
select value, 0
from json_each(sqlc.arg('circle'))
;

-- name: BulkGetCircle :many
select id, name
from circle
where name in (sqlc.slice('circle'))
;

-- name: BulkRemoveCircleFromArchive :exec
delete from archive_circle
where archive_id = ? and circle_id in (select value from json_each(sqlc.arg('circle')))
;


-- name: DecrementCircleCount :exec
update circle
set count = count - 1
where id in (sqlc.slice('circles'))
;

-- name: IncrementCircleCount :exec
update circle
set count = count + 1
where id in (sqlc.slice('circles'))
;

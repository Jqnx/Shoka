-- name: CreateCircle :one
insert into circle (name, count)
values (?, ?)
returning *
;

-- name: GetAllCircle :many
select *
from circle
order by name
;

-- name: GetCircle :one
select *
from circle
where name = ?
;

-- name: GetCircleList :many
select *
from circle
order by name
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
order by artist.name
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

-- name: DeleteCircle :exec
delete from circle
where name = ?
;

-- name: DeleteCircleByID :exec
delete from circle
where id = ?
;

-- name: GetArchivesByCircleName :many
select archive.*, reading_progress.page, reading_progress.last_read, reading_progress.completed
from archive
join archive_circle on archive.id = archive_circle.archive_id
join circle on archive_circle.circle_id = circle.id
left join reading_progress on archive.id = reading_progress.archive_id and reading_progress.user_id = sqlc.arg('uid')
where circle.name = sqlc.arg('name')
  and (sqlc.arg('all_libraries') = 1
       or archive.library_id in (select library_id from library_access where user_id = sqlc.arg('uid')))
order by archive.title asc
limit sqlc.arg('limit')
offset sqlc.arg('offset')
;

-- name: TotalArchiveWithCircle :one
select count(archive.id)
from archive
join archive_circle on archive.id = archive_circle.archive_id
join circle on archive_circle.circle_id = circle.id
where circle.name = sqlc.arg('name')
  and (sqlc.arg('all_libraries') = 1
       or archive.library_id in (select library_id from library_access where user_id = sqlc.arg('uid')))
;

-- name: BulkAddArtistCircle :exec
insert or ignore into artist_circle (artist_id, circle_id)
select a.value, c.value
from json_each(sqlc.arg('artists')) as a, json_each(sqlc.arg('circles')) as c
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

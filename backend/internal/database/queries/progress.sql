-- name: InsertReadingProgress :exec
insert into progress (
  archive_id,
  user_id,
  page,
  completed,
  last_read
) values ( ?, ?, ?, ?, ? )
;

-- name: GetProgressForArchive :one
select *
from progress
where archive_id = ? and user_id = ?
;

-- name: GetAllLastRead :many
select archive_id, last_read
from progress
;

-- name: UpdateReadingProgress :one
update progress
set page = coalesce(sqlc.arg('page'), page),
    completed = coalesce(sqlc.arg('completed'), completed),
    last_read = coalesce(?, last_read)
where archive_id = ? AND user_id = ?
returning
  archive_id,
  user_id,
  page,
  completed,
  last_read
;

-- name: DeleteReadingProgress :exec
delete from progress
where archive_id = ? and user_id = ?
;

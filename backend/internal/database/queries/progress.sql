-- name: InsertReadingProgress :exec
insert into reading_progress (
  archive_id,
  user_id,
  page,
  status,
  last_read
) values ( ?, ?, ?, ?, ? )
;

-- name: GetUserReadingProgress :one
select *
from reading_progress
where archive_id = ? and user_id = ?
;

-- name: UserReadingProgressExists :execresult
select *
from reading_progress
where archive_id = ? and user_id = ?
;

-- name: GetAllLastRead :many
select archive_id, last_read
from reading_progress
;

-- name: UpdateReadingProgress :one
update reading_progress
set page = coalesce(sqlc.arg('page'), page),
    status = coalesce(sqlc.arg('status'), status),
    last_read = coalesce(?, last_read)
where archive_id = ? AND user_id = ?
returning
  archive_id,
  user_id,
  page,
  status,
  last_read
;

-- name: DeleteReadingProgress :exec
delete from reading_progress
where archive_id = ? and user_id = ?
;

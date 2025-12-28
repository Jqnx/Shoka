-- name: InsertReadingProgress :exec
insert into reading_progress (
  archive_id,
  user_id,
  page,
  status,
  last_read
) values ( $1, $2, $3, $4, $5 )
;

-- name: GetUserReadingProgress :one
select *
from reading_progress
where archive_id = $1 and user_id = $2
;

-- name: UserReadingProgressExists :execresult
select *
from reading_progress
where archive_id = $1 and user_id = $2
;

-- name: GetAllLastRead :many
select archive_id, last_read
from reading_progress
;

-- name: UpdateReadingProgress :one
update reading_progress
set page = coalesce(sqlc.narg('page'), page),
    status = coalesce(sqlc.narg('status'), status),
    last_read = coalesce($1, last_read)
where archive_id = $2 AND user_id = $3
returning
  archive_id,
  user_id,
  page,
  status,
  last_read
;

-- name: DeleteReadingProgress :exec
delete from reading_progress
where archive_id = $1 and user_id = $2
;


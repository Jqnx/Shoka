-- name: UpsertReadingProgress :one
-- Idempotent upsert backing PUT /api/archives/{id}/progress. The client
-- reports its current absolute page on every page turn, so an insert-or-
-- update on the (user_id, archive_id) primary key covers both the create
-- and update case with no separate write path needed at the API layer.
insert into reading_progress (archive_id, user_id, page, completed, last_read)
values (?, ?, ?, ?, datetime('now'))
on conflict (user_id, archive_id) do update set
    page = excluded.page,
    completed = excluded.completed,
    last_read = excluded.last_read
returning *
;

-- name: GetProgressForArchive :one
select *
from reading_progress
where archive_id = ? and user_id = ?
;

-- name: DeleteReadingProgress :exec
delete from reading_progress
where archive_id = ? and user_id = ?
;

-- name: GetRecentlyReadArchives :many
-- Deliberately NOT scoped to a library: this is a continue-reading feed
-- spanning the whole collection, regardless of which library each archive
-- belongs to.
select archive.*, reading_progress.page, reading_progress.last_read, reading_progress.completed
from archive
join reading_progress
    on archive.id = reading_progress.archive_id
    and reading_progress.user_id = sqlc.arg('uid')
order by reading_progress.last_read desc
limit sqlc.arg('limit')
offset sqlc.arg('offset')
;

-- name: CountRecentlyReadArchives :one
select count(*)
from reading_progress
where user_id = ?
;

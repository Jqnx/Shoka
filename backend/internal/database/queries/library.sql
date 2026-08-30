-- name: CreateLibrary :one
insert into library (
  id,
  name,
  path,
  type
)
values (?, ?, ?, ?)
returning *;


-- name: GetLibraryByID :one
select *
from library
where id = ?
;

-- name: GetLibraryByPath :one
select *
from library
where path = ?
;

-- name: ListLibraries :many
select *
from library
order by name
;

-- name: UpdateLibrary :one
update library
set
    name = ?,
    scan_interval_minutes = ?,
    watch_enabled = ?,
    updated_at = datetime('now')
where id = ?
returning *;

-- ListLibrariesDueForScan returns every library whose periodic scan is due:
-- scanning is enabled (interval > 0) and either it has never been scanned or
-- the interval has elapsed since the last scan attempt. Polled by the
-- library manager's scheduler.
-- name: ListLibrariesDueForScan :many
select *
from library
where scan_interval_minutes > 0
  and (
    last_scanned_at is null
    or datetime(last_scanned_at, '+' || scan_interval_minutes || ' minutes') <= datetime('now')
  )
order by name;

-- MarkLibraryScanned records a scan attempt. Deliberately does not touch
-- updated_at, which tracks configuration edits rather than scan activity.
-- name: MarkLibraryScanned :exec
update library
set last_scanned_at = datetime('now')
where id = ?;

-- name: DeleteLibrary :exec
delete from library
where id = ?
;

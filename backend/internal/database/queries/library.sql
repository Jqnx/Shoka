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

-- name: ListEnabledLibraries :many
select *
from library
where enabled
order by name
;

-- name: UpdateLibrary :one
update library
set
    name = ?,
    enabled = ?,
    updated_at = datetime('now')
where id = ?
returning *;

-- name: SetLibraryEnabled :one
update library
set
    enabled = ?,
    updated_at = datetime('now')
where id = ?
returning *;

-- name: DeleteLibrary :exec
delete from library
where id = ?
;

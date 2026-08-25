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
    updated_at = datetime('now')
where id = ?
returning *;

-- name: DeleteLibrary :exec
delete from library
where id = ?
;

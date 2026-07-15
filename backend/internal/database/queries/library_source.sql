-- name: GetLibrarySources :many
select *
from library_source
where library_id = ?
order by source
;

-- name: GetLibrarySource :one
select *
from library_source
where library_id = ? and source = ?
;

-- name: UpsertLibrarySource :one
insert into library_source (
  library_id,
  source,
  enabled,
  cookies,
  api_key,
  magazine_blocklist,
  misc_blocklist
)
values (?, ?, ?, ?, ?, ?, ?)
on conflict (library_id, source) do update set
    enabled = excluded.enabled,
    cookies = excluded.cookies,
    api_key = excluded.api_key,
    magazine_blocklist = excluded.magazine_blocklist,
    misc_blocklist = excluded.misc_blocklist
returning *;

-- name: DeleteLibrarySources :exec
delete from library_source
where library_id = ?
;

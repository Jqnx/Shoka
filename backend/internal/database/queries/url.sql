-- name: GetArchiveUrls :many
select url
from archive_url
where archive_id = ?
;

-- name: RemoveArchiveUrl :exec
delete from archive_url
where archive_id = ? and url in (sqlc.slice('urls'))
;

-- name: BulkAddArchiveURLs :exec
insert into archive_url (archive_id, url)
select ?, value from json_each(sqlc.arg('urls'))
where true
on conflict (archive_id, url) do nothing
;

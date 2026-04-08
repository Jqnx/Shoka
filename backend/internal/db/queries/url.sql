-- name: CreateArchiveURL :exec
insert into archive_url (url, archive_id)
values (?, ?)
;

-- name: RemoveArchiveUrl :exec
delete from archive_url
where archive_id = ? and url in (sqlc.slice('urls'))
;

-- name: ArchiveUrlExists :execresult
select url
from archive_url
where url = ?
;

-- name: GetArchiveUrls :many
select id, url
from archive_url
where archive_id = ?
;

-- name: GetArchiveUrlIDs :many
select id
from archive_url
where archive_id = ?
;

-- name: BulkAddArchiveURLs :exec
insert into archive_url (archive_id, url)
select ?, value from json_each(sqlc.arg('urls'))
;

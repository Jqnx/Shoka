-- name: CreateArchiveURL :exec
insert into archive_url (url, archive_id)
values ($1, $2) 
;

-- name: RemoveArchiveUrl :exec
delete from archive_url
where archive_id = $1 and url = any(sqlc.arg('urls')::text[])
;

-- name: ArchiveUrlExists :execresult
select url
from archive_url
where url = $1
;

-- name: GetArchiveUrls :many
select id, url
from archive_url
where archive_id = $1
;

-- name: GetArchiveUrlIDs :many
select id 
from archive_url
where archive_id = $1
;

-- name: BulkAddArchiveURLs :exec
insert into archive_url (archive_id, url)
select $1, unnest(sqlc.arg('urls')::text[])
;

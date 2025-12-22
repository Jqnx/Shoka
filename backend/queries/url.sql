-- name: CreateArchiveURL :exec
insert into archive_url (url, archive_id)
values ($1, $2) 
;

-- name: RemoveArchiveUrl :exec
delete from archive_url
where archive_id = $1
;

-- name: ArchiveUrlExists :execresult
select url
from archive_url
where url = $1
;

-- name: GetArchiveUrls :many
select archive_url.id, archive_url.url
from archive
join archive_url on archive.id = archive_url.archive_id
where archive.id = $1
;


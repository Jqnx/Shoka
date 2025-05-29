-- name: CreateArchiveURL :exec
insert into urls (url, archive_id)
values ($1, $2) 
;

-- name: RemoveArchiveUrl :exec
delete from urls
where archive_id = $1
;

-- name: ArchiveUrlExists :execresult
select url
from urls
where url = $1
;

-- name: GetArchiveURLs :many
select urls.id, urls.url
from archives
join urls on archives.id = urls.archive_id
where archives.archive_id = $1
;


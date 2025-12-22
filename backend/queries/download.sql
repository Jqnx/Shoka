/*
-- name: CreateDownload :one
INSERT INTO downloads (id, url, source, filename, status, progress, error, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
RETURNING *;

-- name: GetDownload :one
select *
from downloads
where id = $1
;

-- name: GetAllDownloads :many
select *
from downloads
order by created_at desc
;

-- name: GetDownloadsByStatus :many
select *
from downloads
where status = any($1::text[])
order by created_at desc
;

-- name: UpdateDownloadStatus :one
UPDATE downloads 
SET status = $2, progress = $3, error = $4, updated_at = $5
WHERE id = $1
RETURNING *;

-- name: UpdateDownloadProgress :one
UPDATE downloads 
SET progress = $2, updated_at = $3
WHERE id = $1
RETURNING *;

-- name: DeleteDownload :exec
delete from downloads
where id = $1
;

-- name: GetPendingDownloads :many
select *
from downloads
where status in ('pending', 'downloading')
order by created_at asc
;
*/


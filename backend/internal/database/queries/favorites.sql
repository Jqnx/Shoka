-- name: AddFavoriteArchive :exec
insert into favorite_archive (archive_id, user_id, favorited_at)
values (?, ?, datetime('now'))
on conflict (archive_id, user_id) do nothing
;

-- name: RemoveFavoriteArchive :exec
delete from favorite_archive
where archive_id = ? and user_id = sqlc.arg('uid')
;

-- name: GetUserFavoriteArchiveList :many
select archive.*, reading_progress.page, reading_progress.last_read, reading_progress.completed
from archive
join favorite_archive as fa on archive.id = fa.archive_id
left join
    reading_progress on archive.id = reading_progress.archive_id and reading_progress.user_id = sqlc.arg('uid')
where fa.user_id = sqlc.arg('uid')
order by fa.favorited_at desc
limit sqlc.arg('limit')
offset sqlc.arg('offset')
;

-- name: CountUserFavoriteArchive :one
select count(*)
from favorite_archive
where user_id = ?
;

-- name: ArchiveIsFavorited :one
select exists (
    select 1 from favorite_archive where archive_id = ? and user_id = sqlc.arg('uid')
) = 1
;

-- name: GetFavoritedArchiveIDs :many
-- Bulk existence check used to attach IsFavorited to archive list responses
-- without an N+1 query per archive.
select archive_id
from favorite_archive
where user_id = sqlc.arg('uid') and archive_id in (sqlc.slice('ids'))
;

-- name: UpsertArchiveRating :one
-- Idempotent upsert backing PUT /api/archives/{id}/rating. Rating is
-- validated (1-5) at the handler layer before this runs.
insert into archive_rating (archive_id, user_id, rating, rated_at)
values (?, ?, ?, datetime('now'))
on conflict (archive_id, user_id) do update set
    rating = excluded.rating,
    rated_at = excluded.rated_at
returning *
;

-- name: RemoveArchiveRating :exec
delete from archive_rating
where archive_id = ? and user_id = sqlc.arg('uid')
;

-- name: GetArchiveRating :one
select rating
from archive_rating
where archive_id = ? and user_id = sqlc.arg('uid')
;

-- name: GetArchiveRatingsForIDs :many
-- Bulk fetch of the current user's own rating for a set of archives, used
-- to attach Rating to archive list responses without an N+1 query per row.
select archive_id, rating
from archive_rating
where user_id = sqlc.arg('uid') and archive_id in (sqlc.slice('ids'))
;

-- name: AddFavoritePage :exec
insert into favorite_page (archive_id, user_id, page_index, favorited_at)
values (sqlc.arg('archive_id'), sqlc.arg('uid'), sqlc.arg('page_index'), datetime('now'))
on conflict (archive_id, user_id, page_index) do nothing
;

-- name: RemoveFavoritePage :exec
delete from favorite_page
where archive_id = sqlc.arg('archive_id')
  and user_id = sqlc.arg('uid')
  and page_index = sqlc.arg('page_index')
;

-- name: GetFavoritePageIndices :many
select page_index
from favorite_page
where archive_id = sqlc.arg('archive_id') and user_id = sqlc.arg('uid')
order by page_index asc
;

-- name: GetUserFavoritePageList :many
-- Cross-library favorite pages, newest first. The page_index < page_count
-- filter hides rows whose archive has since shrunk past that index; the rows
-- are kept so the user can still unfavorite them.
-- fp.rowid is the tiebreaker: favorited_at only has second granularity, and
-- the gallery's global ordinal must line up across separately paginated
-- requests.
select fp.archive_id, fp.page_index, fp.favorited_at,
       a.title, a.page_count
from favorite_page as fp
join archive as a on a.id = fp.archive_id
where fp.user_id = sqlc.arg('uid') and fp.page_index < a.page_count
order by fp.favorited_at desc, fp.rowid desc
limit sqlc.arg('limit')
offset sqlc.arg('offset')
;

-- name: CountUserFavoritePages :one
select count(*)
from favorite_page as fp
join archive as a on a.id = fp.archive_id
where fp.user_id = sqlc.arg('uid') and fp.page_index < a.page_count
;

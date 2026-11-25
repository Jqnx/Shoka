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

-- name: GetUserFavoriteImageList :many
-- Favorite archive pages and audio gallery images across every library the
-- caller may access (all of them for an admin), newest first. Each arm hides
-- rows whose source has since shrunk past the index; the rows are kept so the
-- user can still unfavorite them. favorited_at has second granularity, so the
-- id/index columns break ties and keep the viewer's global ordinal stable
-- across paginated requests.
select kind, archive_id, audio_id, title, item_count, item_index, favorited_at, updated_at
from (
    select 'archive' as kind,
           fp.archive_id as archive_id,
           '' as audio_id,
           a.title as title,
           a.page_count as item_count,
           fp.page_index as item_index,
           fp.favorited_at as favorited_at,
           a.updated_at as updated_at
    from favorite_page as fp
    join archive as a on a.id = fp.archive_id
    where fp.user_id = sqlc.arg('uid') and fp.page_index < a.page_count
      and (sqlc.arg('all_libraries') = 1
           or a.library_id in (select library_id from library_access where user_id = sqlc.arg('uid')))
    union all
    select 'audio' as kind,
           '' as archive_id,
           af.audio_id as audio_id,
           au.title as title,
           (select count(*) from audio_image as ai where ai.audio_id = af.audio_id) as item_count,
           af.image_index as item_index,
           af.favorited_at as favorited_at,
           au.updated_at as updated_at
    from audio_image_favorite as af
    join audio as au on au.id = af.audio_id
    where af.user_id = sqlc.arg('uid')
      and af.image_index < (select count(*) from audio_image as ai where ai.audio_id = af.audio_id)
      and (sqlc.arg('all_libraries') = 1
           or au.library_id in (select library_id from library_access where user_id = sqlc.arg('uid')))
)
order by favorited_at desc, kind asc, archive_id desc, audio_id desc, item_index desc
limit sqlc.arg('limit')
offset sqlc.arg('offset')
;

-- name: CountUserFavoriteImages :one
-- Same library scope and stale-index filters as GetUserFavoriteImageList: the
-- total must describe the same set the list pages through.
select count(*)
from (
    select 1
    from favorite_page as fp
    join archive as a on a.id = fp.archive_id
    where fp.user_id = sqlc.arg('uid') and fp.page_index < a.page_count
      and (sqlc.arg('all_libraries') = 1
           or a.library_id in (select library_id from library_access where user_id = sqlc.arg('uid')))
    union all
    select 1
    from audio_image_favorite as af
    join audio as au on au.id = af.audio_id
    where af.user_id = sqlc.arg('uid')
      and af.image_index < (select count(*) from audio_image as ai where ai.audio_id = af.audio_id)
      and (sqlc.arg('all_libraries') = 1
           or au.library_id in (select library_id from library_access where user_id = sqlc.arg('uid')))
)
;

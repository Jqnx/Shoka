-- name: CreateArtist :one
insert into artist (name, count)
values (?, ?)
returning *
;

-- name: CreateAlias :exec
insert into artist_alias (alias, artist_id)
values (?, ?)
;

-- name: CreateArtistUrl :exec
insert into artist_url (url, artist_id)
values (?, ?)
;


-- name: EnsureArtistExist :many
INSERT INTO artist (name, count)
SELECT value, 0
FROM json_each(?)
ON CONFLICT (name) DO UPDATE
SET name = excluded.name
RETURNING id, name
;


-- name: GetAllArtists :many
select *
from artist
order by name
;

-- name: GetArtistByName :one
select *
from artist
where name = ?
;

-- name: GetArtistByID :one
select *
from artist
where id = ?
;

-- name: GetArtistList :many
select artist.id,
       artist.name,
       artist.count,
       cover.id as cover_archive_id,
       cover.cover_page as cover_page,
       cover.page_count as cover_page_count
from artist
left join archive as cover on cover.id = coalesce(
    -- A stored choice only counts while that archive still exists and the
    -- caller can read its library.
    (select chosen.id from archive chosen
     where chosen.id = artist.cover_archive_id
       and (cast(sqlc.arg('cover_all_libs') as integer) = 1
            or chosen.library_id in (select library_access.library_id from library_access where library_access.user_id = sqlc.arg('uid')))),
    -- Default: the artist's oldest readable archive (created_at tie-broken
    -- by id), so the card image does not lurch around as the library grows.
    (select aa.archive_id
     from archive_artist aa
     join archive oldest on oldest.id = aa.archive_id
     where aa.artist_id = artist.id
       and (cast(sqlc.arg('cover_all_libs') as integer) = 1
            or oldest.library_id in (select library_access.library_id from library_access where library_access.user_id = sqlc.arg('uid')))
     order by oldest.created_at, oldest.id
     limit 1)
)
where (cast(sqlc.arg('favorites_only') as integer) = 0
       or exists (select 1 from artist_favorite af
                  where af.artist_id = artist.id and af.user_id = sqlc.arg('uid')))
  and (cast(sqlc.arg('filter_rating') as integer) = 0
       or exists (select 1 from artist_rating ar
                  where ar.artist_id = artist.id
                    and ar.user_id = sqlc.arg('uid')
                    and ar.rating = cast(sqlc.arg('filter_rating') as integer)))
order by artist.name
limit sqlc.arg('limit')
offset sqlc.arg('offset')
;

-- name: GetArtistCoverArchive :one
select cover.id as archive_id, cover.cover_page, cover.page_count
from artist
left join archive as cover on cover.id = coalesce(
    (select chosen.id from archive chosen
     where chosen.id = artist.cover_archive_id
       and (cast(sqlc.arg('cover_all_libs') as integer) = 1
            or chosen.library_id in (select library_access.library_id from library_access where library_access.user_id = sqlc.arg('uid')))),
    (select aa.archive_id
     from archive_artist aa
     join archive oldest on oldest.id = aa.archive_id
     where aa.artist_id = artist.id
       and (cast(sqlc.arg('cover_all_libs') as integer) = 1
            or oldest.library_id in (select library_access.library_id from library_access where library_access.user_id = sqlc.arg('uid')))
     order by oldest.created_at, oldest.id
     limit 1)
)
where artist.id = cast(sqlc.arg('artist_id') as integer)
;

-- name: GetArtistCircles :many
select circle.id, circle.name
from artist_circle
join circle on circle.id = artist_circle.circle_id
where artist_circle.artist_id = cast(sqlc.arg('artist_id') as integer)
order by circle.name
;

-- name: ArchiveArtistLinkExists :one
select exists (
    select 1 from archive_artist
    where archive_id = cast(sqlc.arg('archive_id') as text)
      and artist_id = cast(sqlc.arg('artist_id') as integer)
) = 1
;

-- name: UpdateArtistCover :exec
update artist
set cover_archive_id = ?
where id = ?
;

-- name: AddFavoriteArtist :exec
insert into artist_favorite (artist_id, user_id, favorited_at)
values (?, ?, datetime('now'))
on conflict (artist_id, user_id) do nothing
;

-- name: RemoveFavoriteArtist :exec
delete from artist_favorite
where artist_id = ? and user_id = sqlc.arg('uid')
;

-- name: ArtistIsFavorited :one
select exists (
    select 1 from artist_favorite where artist_id = ? and user_id = sqlc.arg('uid')
) = 1
;

-- name: GetFavoritedArtistIDs :many
select artist_id
from artist_favorite
where user_id = sqlc.arg('uid') and artist_id in (sqlc.slice('ids'))
;

-- name: UpsertArtistRating :one
insert into artist_rating (artist_id, user_id, rating, rated_at)
values (?, ?, ?, datetime('now'))
on conflict (artist_id, user_id) do update set
    rating = excluded.rating,
    rated_at = excluded.rated_at
returning *
;

-- name: RemoveArtistRating :exec
delete from artist_rating
where artist_id = ? and user_id = sqlc.arg('uid')
;

-- name: GetArtistRating :one
select rating
from artist_rating
where artist_id = ? and user_id = sqlc.arg('uid')
;

-- name: GetArtistRatingsForIDs :many
select artist_id, rating
from artist_rating
where user_id = sqlc.arg('uid') and artist_id in (sqlc.slice('ids'))
;

-- name: GetArtistUrls :many
select artist_url.id, artist_url.url
from artist
join artist_url on artist.id = artist_url.artist_id
where name = ?
;

-- name: GetArtistAliases :many
select artist_alias.id, artist_alias.alias
from artist
join artist_alias on artist.id = artist_alias.artist_id
where name = ?
;

-- name: GetArtistUrlsByID :many
select id, url
from artist_url
where artist_id = ?
order by url
;

-- name: GetArtistAliasesByID :many
select id, alias
from artist_alias
where artist_id = ?
order by alias
;

-- name: ArtistExists :execresult
select name
from artist
where name = ?
;

-- name: ArtistAliasExists :execresult
select alias
from artist_alias
where alias = ?
;

-- name: ArtistUrlExists :execresult
select url
from artist_url
where url = ?
;

-- name: TotalArtists :one
select count(artist.id)
from artist
where (cast(sqlc.arg('favorites_only') as integer) = 0
       or exists (select 1 from artist_favorite af
                  where af.artist_id = artist.id and af.user_id = sqlc.arg('uid')))
  and (cast(sqlc.arg('filter_rating') as integer) = 0
       or exists (select 1 from artist_rating ar
                  where ar.artist_id = artist.id
                    and ar.user_id = sqlc.arg('uid')
                    and ar.rating = cast(sqlc.arg('filter_rating') as integer)))
;

-- name: TotalArchiveWithArtist :one
-- Same library scope as GetArchivesByArtistID: the total must describe the
-- same set the list pages through.
select count(archive.id)
from archive
join archive_artist on archive.id = archive_artist.archive_id
join artist on archive_artist.artist_id = artist.id
where artist.id = sqlc.arg('id')
  and (sqlc.arg('all_libraries') = 1
       or archive.library_id in (select library_access.library_id from library_access where library_access.user_id = sqlc.arg('uid')))
;

-- name: AddArtistToArchive :exec
insert into archive_artist (archive_id, artist_id)
values (?, ?)
;

-- name: UpdateArtist :one
update artist
set name = ?
where name = sqlc.arg('old_name')
returning *
;

-- name: UpdateArtistByID :one
update artist
set name = ?
where id = sqlc.arg('id')
returning *
;

-- name: DecrementArtistCount :exec
update artist
set count = count - 1
where id in (sqlc.slice('artists'))
;

-- name: IncrementArtistCount :exec
update artist
set count = count + 1
where id in (sqlc.slice('artists'))
;


-- name: RemoveArtistAliases :exec
delete from artist_alias
where artist_id = ?
;

-- name: RemoveArtistUrls :exec
delete from artist_url
where artist_id = ?
;

-- name: DeleteArtist :exec
delete from artist
where id = ?
;

-- name: DeleteAllArtist :exec
delete from artist
;

-- name: GetArchivesByArtistID :many
select archive.*, reading_progress.page, reading_progress.last_read, reading_progress.completed
from archive
join archive_artist on archive.id = archive_artist.archive_id
join artist on archive_artist.artist_id = artist.id
left join reading_progress on archive.id = reading_progress.archive_id and reading_progress.user_id = sqlc.arg('uid')
where artist.id = sqlc.arg('id')
  and (sqlc.arg('all_libraries') = 1
       or archive.library_id in (select library_access.library_id from library_access where library_access.user_id = sqlc.arg('uid')))
order by archive.title asc
limit sqlc.arg('limit')
offset sqlc.arg('offset')
;

-- name: GetArchiveArtists :many
select artist.*
from archive
join archive_artist on archive.id = archive_artist.archive_id
join artist on archive_artist.artist_id = artist.id
where archive.id = ?
;

-- name: GetArchiveArtistIDs :many
select artist.id
from archive
join archive_artist on archive.id = archive_artist.archive_id
join artist on archive_artist.artist_id = artist.id
where archive.id = ?
;

-- name: GetArchiveByArtist :many
select archive.*
from archive
join archive_artist on archive.id = archive_artist.archive_id
join artist on archive_artist.artist_id = artist.id
where artist.name = ?
;

-- name: GetArchiveByArtistList :many
select archive.*
from archive
join archive_artist on archive.id = archive_artist.archive_id
join artist on archive_artist.artist_id = artist.id
where artist.name = ?
limit ?
offset ?
;

-- name: GetArchiveIDsByArtist :many
select archive.id
from archive
join archive_artist on archive.id = archive_artist.archive_id
join artist on archive_artist.artist_id = artist.id
where artist.name = ?
;

-- name: GetArchiveIDsByArtists :many
select archive.id
from archive
join archive_artist on archive.id = archive_artist.archive_id
join artist on archive_artist.artist_id = artist.id
where artist.name in sqlc.slice('artists')
group by archive.id
having count(distinct artist.id) = sqlc.arg('amount')
;

-- name: BulkAddArchiveArtists :exec
insert or ignore into archive_artist (archive_id, artist_id)
select ?, value
from json_each(sqlc.arg('artists'))
;

-- name: BulkAddArtists :exec
insert or ignore into artist (name, count)
select value, 0
from json_each(sqlc.arg('artists'))
;

-- name: BulkGetArtists :many
select id, name
from artist
where name in (sqlc.slice('artists'))
;

-- name: BulkRemoveArtistFromArchive :exec
delete from archive_artist
where archive_id = ? and artist_id in (select value from json_each(sqlc.arg('artists')))
;

-- name: GetArtistAliasTargets :many
-- Every alias and the canonical name of the artist that owns it, so the
-- caller can fold keys with Go's Unicode-aware lower-casing instead of
-- SQLite's ASCII-only lower(). An alias shadowed by an artist whose exact
-- name equals the alias is ignored: a real artist row outranks an alias.
-- Ordered by alias id so the caller's first-row-per-folded-key choice is
-- deterministic.
select artist_alias.alias, artist.name as artist_name
from artist_alias
join artist on artist.id = artist_alias.artist_id
where not exists (select 1 from artist named where named.name = trim(artist_alias.alias))
order by artist_alias.id
;

-- name: GetArtistNamesByNames :many
-- The exact artist names among the submitted names. Used to keep an exact
-- artist name from being rewritten by another artist's case-variant alias.
select name
from artist
where name in (select value from json_each(cast(sqlc.arg('names_json') as text)))
;

-- name: GetArtistAliasOwnership :many
-- Every alias with the artist that owns it. The caller folds the alias with
-- Go's Unicode-aware lower-casing so merge conflict detection matches the
-- alias resolver instead of SQLite's ASCII-only lower().
select artist_id, alias
from artist_alias
order by id
;

-- name: RepointArtistArchiveLinks :exec
insert or ignore into archive_artist (archive_id, artist_id)
select archive_id, cast(sqlc.arg('target_id') as integer)
from archive_artist
where artist_id = cast(sqlc.arg('source_id') as integer)
;

-- name: MoveArtistUrls :exec
update artist_url
set artist_id = cast(sqlc.arg('target_id') as integer)
where artist_id = cast(sqlc.arg('source_id') as integer)
;

-- name: CopyArtistCircles :exec
insert or ignore into artist_circle (artist_id, circle_id)
select cast(sqlc.arg('target_id') as integer), circle_id
from artist_circle
where artist_id = cast(sqlc.arg('source_id') as integer)
;

-- name: CopyArtistFavorites :exec
insert or ignore into artist_favorite (artist_id, user_id, favorited_at)
select cast(sqlc.arg('target_id') as integer), user_id, favorited_at
from artist_favorite
where artist_id = cast(sqlc.arg('source_id') as integer)
;

-- name: CopyArtistRatings :exec
insert or ignore into artist_rating (artist_id, user_id, rating, rated_at)
select cast(sqlc.arg('target_id') as integer), user_id, rating, rated_at
from artist_rating
where artist_id = cast(sqlc.arg('source_id') as integer)
;

-- name: AdoptArtistCover :exec
update artist
set cover_archive_id = (select cover_archive_id from artist where id = cast(sqlc.arg('source_id') as integer))
where id = cast(sqlc.arg('target_id') as integer)
  and cover_archive_id is null
  and (select cover_archive_id from artist where id = cast(sqlc.arg('source_id') as integer)) is not null
;

-- name: RecomputeArtistCount :exec
update artist
set count = (select count(*) from archive_artist where artist_id = cast(sqlc.arg('count_id') as integer))
          + (select count(*) from audio_artist where artist_id = cast(sqlc.arg('count_id') as integer))
where id = cast(sqlc.arg('count_id') as integer)
;

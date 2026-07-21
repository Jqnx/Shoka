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
select *
from artist
order by name
limit ?
offset ?
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
select count(id)
from artist
;

-- name: TotalArchiveWithArtist :one
select count(archive.id)
from archive
join archive_artist on archive.id = archive_artist.archive_id
join artist on archive_artist.artist_id = artist.id
where artist.id = ?
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

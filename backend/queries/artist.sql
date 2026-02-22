-- name: CreateArtist :one
insert into artist (name, count)
values ($1, $2)
returning *
;

-- name: CreateAlias :exec
insert into artist_alias (alias, artist_id)
values ($1, $2)
;

-- name: CreateArtistUrl :exec
insert into artist_url (url, artist_id)
values ($1, $2)
;

-- name: BulkAddArchiveArtists :exec
insert into archive_artist (archive_id, artist_id)
select $1, unnest(sqlc.arg('artists')::int[])
;

-- name: EnsureArtistExist :many
insert into artist (name, count)
select unnest(sqlc.arg('artists')::text[]), 0
on conflict (name) do update
set name = excluded.name
returning id, name
;

-- name: GetAllArtists :many
select id, name, count
from artist
order by id
;

-- name: GetArtistByName :one
select id, name, count
from artist
where name = $1
;

-- name: GetArtistList :many
select id, name, count
from artist
order by $1
limit $2
offset $3
;

-- name: GetArtistUrls :many
select artist_url.id, artist_url.url
from artist
join artist_url on artist.id = artist_url.artist_id
where name = $1
;

-- name: GetArtistAliases :many
select artist_alias.id, artist_alias.alias
from artist
join artist_alias on artist.id = artist_alias.artist_id
where name = $1
;

-- name: ArtistExists :execresult
select name
from artist
where name = $1
;

-- name: ArtistAliasExists :execresult
select alias
from artist_alias
where alias = $1
;

-- name: ArtistUrlExists :execresult
select url
from artist_url
where url = $1
;

-- name: TotalArtists :one
select count(id)
from artist
;

-- name: AddArtistToArchive :exec
insert into archive_artist (archive_id, artist_id)
values ($1, $2)
;

-- name: UpdateArtist :one
update artist
set name = $1
where name = sqlc.arg(old_name)::text
returning *
;

-- name: DecrementArtistCount :exec
update artist
set count = count - 1
where id = any(sqlc.arg('artists')::int[])
;

-- name: IncrementArtistCount :exec
update artist
set count = count + 1
where id = any(sqlc.arg('artists')::int[])
;


-- name: RemoveArtistFromArchive :exec
delete from archive_artist
where archive_id = $1 and artist_id = any(sqlc.arg('artists')::int[])
;

-- name: RemoveArtistAliases :exec
delete from artist_alias
where artist_id = $1
;

-- name: RemoveArtistUrls :exec
delete from artist_url
where artist_id = $1
;

-- name: DeleteArtist :exec
delete from artist
where id = $1
;

-- name: DeleteAllArtist :exec
delete from artist
;

-- name: GetArchiveArtists :many
select artist.*
from archive
join archive_artist on archive.id = archive_artist.archive_id
join artist on archive_artist.artist_id = artist.id
where archive.id = $1
;

-- name: GetArchiveArtistIDs :many
select artist.id
from archive
join archive_artist on archive.id = archive_artist.archive_id
join artist on archive_artist.artist_id = artist.id
where archive.id = $1
;

-- name: GetArchiveByArtist :many
select
    archive.id,
    archive.title,
    archive.summary,
    archive.language,
    archive.category,
    archive.page_count,
    archive.file_path,
    archive.file_hash,
    archive.thumb_path,
    archive.created_at,
    archive.updated_at,
    archive.release_date
from archive
join archive_artist on archive.id = archive_artist.archive_id
join artist on archive_artist.artist_id = artist.id
where artist.name = $1
;

-- name: GetArchiveByArtistList :many
select
    archive.id,
    archive.title,
    archive.summary,
    archive.language,
    archive.category,
    archive.page_count,
    archive.file_path,
    archive.file_hash,
    archive.thumb_path,
    archive.created_at,
    archive.updated_at,
    archive.release_date
from archive
join archive_artist on archive.id = archive_artist.archive_id
join artist on archive_artist.artist_id = artist.id
where artist.name = $1
limit $2
offset $3
;

-- name: GetArchiveIDsByArtist :many
select archive.id
from archive
join archive_artist on archive.id = archive_artist.archive_id
join artist on archive_artist.artist_id = artist.id
where artist.name = $1
;

-- name: GetArchiveIDsByArtists :many
select archive.id
from archive
join archive_artist on archive.id = archive_artist.archive_id
join artist on archive_artist.artist_id = artist.id
where artist.name = any(sqlc.arg('artists')::text[])
group by archive.id
having count(distinct artist.id) = sqlc.arg('amount')
;

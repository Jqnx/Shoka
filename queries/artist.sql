-- name: CreateArtist :one
insert into artists (name, count, created_at, updated_at)
values ($1, $2, $3, $4)
returning *
;

-- name: CreateAlias :exec
insert into artist_aliases (alias, artist_id)
values ($1, $2)
;

-- name: CreateArtistLink :exec
insert into artist_links (link, artist_id)
values ($1, $2)
;

-- name: GetAllArtists :many
select id, name, count, created_at, updated_at
from artists
order by id
;

-- name: GetArtistByName :one
select id, name, count, created_at, updated_at
from artists
where name = $1
;

-- name: GetArtistList :many
select id, name, count, created_at, updated_at
from artists
order by $1
limit $2
offset $3
;

-- name: GetArtistLinks :many
select artist_links.id, artist_links.link
from artists
join artist_links on artists.id = artist_links.artist_id
where name = $1
;

-- name: GetArtistAliases :many
select artist_aliases.id, artist_aliases.alias
from artists
join artist_aliases on artists.id = artist_aliases.artist_id
where name = $1
;

-- name: GetArtistGroups :many
select groups.id, groups.name
from artists
join artists_groups on artists.id = artists_groups.artist_id
join groups on artists_groups.group_id = groups.id
where artists.name = $1
;

-- name: ArtistExists :execresult
select name
from artists
where name = $1
;

-- name: ArtistAliasExists :execresult
select alias
from artist_aliases
where alias = $1
;

-- name: ArtistLinkExists :execresult
select link
from artist_links
where link = $1
;

-- name: TotalArtists :one
select count(id)
from artists
;

-- name: AddArtistToGroup :exec
insert into artists_groups (artist_id, group_id)
values ($1, $2)
;

-- name: AddArtistToArchive :exec
insert into archives_artists (archive_id, artist_id)
values ($1, $2)
;

-- name: UpdateArtist :one
update artists
set name = $1,
    updated_at = $2
where name = sqlc.arg(old_name)::text
returning *
;

-- name: UpdateArtistCount :exec
update artists
set count = $1
where id = $2
;

-- name: RemoveArtistFromArchive :many
delete from archives_artists
where archive_id = $1
returning
    artist_id,
    (select artists.count from artists where artists.id = archives_artists.artist_id)
;

-- name: RemoveArtistAliases :exec
delete from artist_aliases
where artist_id = $1
;

-- name: RemoveArtistLinks :exec
delete from artist_links
where artist_id = $1
;

-- name: RemoveArtistFromGroup :exec
delete from artists_groups
where artist_id = $1
;

-- name: DeleteArtist :exec
delete from artists
where id = $1
;

-- name: GetArchiveArtists :many
select artists.*
from archives
join archives_artists on archives.id = archives_artists.archive_id
join artists on archives_artists.artist_id = artists.id
where archives.archive_id = $1
;


-- name: CreateArchive :one
insert into archives (
    title,
    summary,
    lang,
    category,
    page_count,
    file_path,
    a_id,
    created_at,
    updated_at
    )
values ($1, $2, $3, $4, $5, $6, $7, $8, $9)
returning *
;

-- name: CreateArchiveURL :exec
insert into urls (url, archive_id)
values ($1, $2) 
;


-- name: GetArchiveLastAID :one
select a_id
from archives
order by a_id desc
limit 1
;

-- name: GetArchiveByAID :one
select *
from archives
where a_id = $1
;

-- name: GetArchivesByTag :many
select a_id
from archives
join archives_tags on archives.id = archives_tags.archive_id
join tags on archives_tags.tag_id = tags.id
where tags.tag = $1
;

-- name: GetArchivesByCharacter :many
select a_id
from archives
join archives_characters on archives.id = archives_characters.archive_id
join characters on archives_characters.character_id = characters.id
where characters.character = $1
;

-- name: GetArchivesByParody :many
select a_id
from archives
join archives_parodies on archives.id = archives_parodies.archive_id
join parodies on archives_parodies.parody_id = parodies.id
where parodies.parody = $1
;

-- name: GetAllArchives :many
select *
from archives
order by a_id
;

-- name: GetArchiveTags :many
select tags.tag
from archives
join archives_tags on archives.id = archives_tags.archive_id
join tags on archives_tags.tag_id = tags.id
where archives.a_id = $1
;

-- name: GetArchiveCharacters :many
select characters.character
from archives
join archives_characters on archives.id = archives_characters.archive_id
join characters on archives_characters.character_id = characters.id
where archives.a_id = $1
;

-- name: GetArchiveParodies :many
select parodies.parody
from archives
join archives_parodies on archives.id = archives_parodies.archive_id
join parodies on archives_parodies.parody_id = parodies.id
where archives.a_id = $1
;

-- name: GetArchiveURLs :many
select urls.url
from archives
join urls on archives.id = urls.archive_id
where archives.a_id = $1
;

-- name: GetArchiveArtists :many
select artists.name
from archives
join archives_artists on archives.id = archives_artists.archive_id
join artists on archives_artists.artist_id = artists.id
where archives.a_id = $1
;

-- name: ArchiveUrlExists :execresult
select *
from urls
where url = $1
;

-- name: ArchiveExists :execresult
select *
from archives
where title = $1
;

-- name: FilePathExists :execresult
select *
from archives
where file_path = $1
;

-- name: ArchiveAIDExists :execresult
select *
from archives
where a_id = $1
;

-- name: UpdateArchive :one
update archives
set title = $1,
    summary = $2,
    lang = $3,
    category = $4,
    page_count = $5,
    file_path = $6,
    updated_at = $7
where a_id = $8
returning *
;

-- name: RemoveArchiveUrl :exec
delete from urls
where archive_id = $1
;

-- name: DeleteArchive :exec
delete from archives
where a_id = $1
;


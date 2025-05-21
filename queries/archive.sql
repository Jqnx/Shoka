-- name: CreateArchive :one
insert into archives (
    title,
    summary,
    language,
    category,
    page_count,
    file_path,
    archive_id,
    hash,
    thumbs_path,
    cover_path,
-- pages_path,
    type,
    created_at,
    updated_at
    )
values ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
returning *
;

-- name: CreateArchiveURL :exec
insert into urls (url, archive_id)
values ($1, $2) 
;


-- name: GetLastArchiveID :one
select archive_id
from archives
order by archive_id desc
limit 1
;

-- name: GetArchiveByID :one
select *
from archives
where archive_id = $1
;

-- name: GetArchiveByFilePath :one
select *
from archives
where file_path = $1
;

-- name: GetArchivesByTag :many
select archives.archive_id
from archives
join archives_tags on archives.id = archives_tags.archive_id
join tags on archives_tags.tag_id = tags.id
where tags.tag = $1
;

-- name: GetArchivesByCharacter :many
select archives.archive_id
from archives
join archives_characters on archives.id = archives_characters.archive_id
join characters on archives_characters.character_id = characters.id
where characters.character = $1
;

-- name: GetArchivesByParody :many
select archives.archive_id
from archives
join archives_parodies on archives.id = archives_parodies.archive_id
join parodies on archives_parodies.parody_id = parodies.id
where parodies.parody = $1
;

-- name: GetAllArchives :many
select *
from archives
order by archive_id
;

-- name: GetArchiveList :many
select *
from archives
limit $1
offset $2
;

-- name: GetArchiveTags :many
select tags.id, tags.tag
from archives
join archives_tags on archives.id = archives_tags.archive_id
join tags on archives_tags.tag_id = tags.id
where archives.archive_id = $1
;

-- name: GetArchiveCharacters :many
select characters.id, characters.character
from archives
join archives_characters on archives.id = archives_characters.archive_id
join characters on archives_characters.character_id = characters.id
where archives.archive_id = $1
;

-- name: GetArchiveParodies :many
select parodies.id, parodies.parody
from archives
join archives_parodies on archives.id = archives_parodies.archive_id
join parodies on archives_parodies.parody_id = parodies.id
where archives.archive_id = $1
;

-- name: GetArchiveURLs :many
select urls.id, urls.url
from archives
join urls on archives.id = urls.archive_id
where archives.archive_id = $1
;

-- name: GetArchiveArtists :many
select artists.id, artists.name
from archives
join archives_artists on archives.id = archives_artists.archive_id
join artists on archives_artists.artist_id = artists.id
where archives.archive_id = $1
;

-- name: ArchiveUrlExists :execresult
select url
from urls
where url = $1
;

-- name: ArchiveExists :execresult
select title
from archives
where title = $1
;

-- name: FilePathExists :execresult
select file_path
from archives
where file_path = $1
;

-- name: ArchiveIDExists :execresult
select archive_id
from archives
where archive_id = $1
;

-- name: ThumbsPathExistsForFilePath :execresult
select thumbs_path
from archives
where file_path = $1
;

-- name: UpdateArchive :one
update archives
set title = $1,
    summary = $2,
    language = $3,
    category = $4,
    page_count = $5,
    file_path = $6,
    updated_at = $7
where archive_id = $8
returning *
;

-- name: UpdateThumbPath :exec
update archives
set thumbs_path = $1
where archive_id = $2
;

-- name: UpdateCoverPath :exec
update archives
set cover_path = $1
where archive_id = $2
;

-- name: RemoveArchiveUrl :exec
delete from urls
where archive_id = $1
;

-- name: DeleteArchive :exec
delete from archives
where archive_id = $1
;


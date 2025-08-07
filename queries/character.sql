-- name: CreateCharacter :one
insert into characters (name, count)
values ($1, $2)
returning *
;

-- name: AddCharacterToArchive :exec
insert into archives_characters (archive_id, character_id)
values ($1, $2)
;

-- name: GetCharacter :one
select id, name, count
from characters
where name = $1
;

-- name: GetAllCharacter :many
select id, name, count
from characters
order by id
;

-- name: CharacterExists :execresult
select id, name
from characters
where name = $1
;

-- name: RemoveCharacterFromArchive :many
delete from archives_characters
where archive_id = $1
returning
    character_id,
    (
        select characters.count
        from characters
        where characters.id = archives_characters.character_id
    )
;

-- name: GetArchiveCharacters :many
select characters.id, characters.name, characters.count
from archives
join archives_characters on archives.id = archives_characters.archive_id
join characters on archives_characters.character_id = characters.id
where archives.archive_id = $1
;

-- name: GetArchivesByCharacter :many
select
    archives.id,
    archives.title,
    archives.summary,
    archives.language,
    archives.category,
    archives.page_count,
    archives.file_path,
    archives.archive_id,
    archives.hash,
    archives.thumbs_path,
    archives.cover_path,
    archives.type,
    archives.created_at,
    archives.updated_at,
    archives.release_date,
    reading_progress.page
from archives
join archives_characters on archives.id = archives_characters.archive_id
join characters on archives_characters.character_id = characters.id
left join
    reading_progress
    on archives.id = reading_progress.archive_id
    and reading_progress.user_id = sqlc.arg('uid')
where characters.name = $1
;

-- name: GetArchiveIDsByCharacter :many
select archives.archive_id
from archives
join archives_characters on archives.id = archives_characters.archive_id
join characters on archives_characters.character_id = characters.id
where characters.name = $1
;

-- name: GetArchivesByCharacterList :many
select
    archives.id,
    archives.title,
    archives.summary,
    archives.language,
    archives.category,
    archives.page_count,
    archives.file_path,
    archives.archive_id,
    archives.hash,
    archives.thumbs_path,
    archives.cover_path,
    archives.type,
    archives.created_at,
    archives.updated_at,
    archives.release_date,
    reading_progress.page
from archives
join archives_characters on archives.id = archives_characters.archive_id
join characters on archives_characters.character_id = characters.id
left join
    reading_progress
    on archives.id = reading_progress.archive_id
    and reading_progress.user_id = sqlc.arg('uid')
where characters.name = $1
limit $2
offset $3
;

-- name: TotalArchivesWithCharacter :one
select count(archives.archive_id)
from archives
join archives_characters on archives.id = archives_characters.archive_id
join characters on archives_characters.character_id = characters.id
where characters.name = $1
;

-- name: UpdateCharacterCount :exec
update characters
set count = $1
where id = $2
;


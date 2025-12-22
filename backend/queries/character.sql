-- name: CreateCharacter :one
insert into character (name, count)
values ($1, $2)
returning *
;

-- name: AddCharacterToArchive :exec
insert into archive_character (archive_id, character_id)
values ($1, $2)
;

-- name: GetCharacter :one
select id, name, count
from character
where name = $1
;

-- name: GetAllCharacter :many
select id, name, count
from character
order by id
;

-- name: CharacterExists :execresult
select id, name
from character
where name = $1
;

-- name: RemoveCharacterFromArchive :many
delete from archive_character
where archive_id = $1
returning
    character_id,
    (
        select character.count
        from character
        where character.id = archive_character.character_id
    )
;

-- name: GetArchiveCharacters :many
select character.id, character.name, character.count
from archive
join archive_character on archive.id = archive_character.archive_id
join character on archive_character.character_id = character.id
where archive.id = $1
;

-- name: GetArchiveByCharacter :many
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
    archive.release_date,
    reading_progress.page
from archive
join archive_character on archive.id = archive_character.archive_id
join character on archive_character.character_id = character.id
left join
    reading_progress
    on archive.id = reading_progress.archive_id
    and reading_progress.user_id = sqlc.arg('uid')
where character.name = $1
;

-- name: GetArchiveIDsByCharacter :many
select archive.id
from archive
join archive_character on archive.id = archive_character.archive_id
join character on archive_character.character_id = character.id
where character.name = $1
;

-- name: GetArchiveByCharacterList :many
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
    archive.release_date,
    reading_progress.page
from archive
join archive_character on archive.id = archive_character.archive_id
join character on archive_character.character_id = character.id
left join
    reading_progress
    on archive.id = reading_progress.archive_id
    and reading_progress.user_id = sqlc.arg('uid')
where character.name = $1
limit $2
offset $3
;

-- name: TotalArchiveWithCharacter :one
select count(archive.id)
from archive
join archive_character on archive.id = archive_character.archive_id
join character on archive_character.character_id = character.id
where character.name = $1
;

-- name: UpdateCharacterCount :exec
update character
set count = $1
where id = $2
;


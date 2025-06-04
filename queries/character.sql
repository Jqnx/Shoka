-- name: CreateCharacter :one
insert into characters (character)
values ($1)
returning *
;

-- name: AddCharacterToArchive :exec
insert into archives_characters (archive_id, character_id)
values ($1, $2)
;

-- name: GetCharacter :one
select id, character
from characters
where character = $1
;

-- name: GetAllCharacter :many
select id, character
from characters
order by id
;

-- name: CharacterExists :execresult
select id, character
from characters
where character = $1
;

-- name: RemoveCharacterFromArchive :exec
delete from archives_characters
where archive_id = $1
;

-- name: GetArchiveCharacters :many
select characters.id, characters.character
from archives
join archives_characters on archives.id = archives_characters.archive_id
join characters on archives_characters.character_id = characters.id
where archives.archive_id = $1
;

-- name: GetArchivesByCharacter :many
select archives.archive_id
from archives
join archives_characters on archives.id = archives_characters.archive_id
join characters on archives_characters.character_id = characters.id
where characters.character = $1
;


-- name: GetArchivesByCharacterList :many
select archives.*
from archives
join archives_characters on archives.id = archives_characters.archive_id
join characters on archives_characters.character_id = characters.id
where characters.character = $1
limit $2
offset $3
;


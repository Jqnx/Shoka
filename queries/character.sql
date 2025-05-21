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


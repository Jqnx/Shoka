-- name: CreateCharacter :one
insert into character (name, count)
values (?, ?)
returning *
;

-- name: AddCharacterToArchive :exec
insert into archive_character (archive_id, character_id)
values (?, ?)
;


-- name: EnsureCharacterExist :many
insert into character (name, count)
select value, 0 from json_each(sqlc.arg('characters'))
on conflict (name) do update
set name = excluded.name
returning id, name
;

-- name: GetCharacter :one
select *
from character
where name = ?
;

-- name: GetAllCharacter :many
select *
from character
order by id
;

-- name: CharacterExists :execresult
select id, name
from character
where name = ?
;

-- name: RemoveCharacterFromArchive :exec
delete from archive_character
where archive_id = ? and character_id in (sqlc.slice('characters'))
;

-- name: GetArchiveCharacters :many
select character.*
from archive
join archive_character on archive.id = archive_character.archive_id
join character on archive_character.character_id = character.id
where archive.id = ?
;

-- name: GetArchiveCharacterIDs :many
select character.id
from archive
join archive_character on archive.id = archive_character.archive_id
join character on archive_character.character_id = character.id
where archive.id = ?
;

-- name: GetArchiveByCharacter :many
select archive.*, progress.page
from archive
join archive_character on archive.id = archive_character.archive_id
join character on archive_character.character_id = character.id
left join
    progress on archive.id = progress.archive_id and progress.user_id = sqlc.arg('uid')
where character.name = ?
;

-- name: GetArchiveIDsByCharacter :many
select archive.id
from archive
join archive_character on archive.id = archive_character.archive_id
join character on archive_character.character_id = character.id
where character.name = ?
;

-- name: GetArchiveIDsByCharacters :many
select archive.id
from archive
join archive_character on archive.id = archive_character.archive_id
join character on archive_character.character_id = character.id
where character.name in (sqlc.slice('characters'))
group by archive.id
having count(distinct character.id) = sqlc.arg('amount')
;

-- name: GetArchiveByCharacterList :many
select archive.*, progress.page
from archive
join archive_character on archive.id = archive_character.archive_id
join character on archive_character.character_id = character.id
left join
    progress on archive.id = progress.archive_id and progress.user_id = sqlc.arg('uid')
where character.name = ?
order by
    case when sqlc.arg('order_by') = 'title_asc' then archive.title end asc,
    case when sqlc.arg('order_by') = 'title_desc' then archive.title end desc,
    case when sqlc.arg('order_by') = 'page_count_asc' then archive.page_count end asc,
    case when sqlc.arg('order_by') = 'page_count_desc' then archive.page_count end desc,
    case when sqlc.arg('order_by') = 'created_at_asc' then archive.created_at end asc,
    case when sqlc.arg('order_by') = 'created_at_desc' then archive.created_at end desc,
    case when sqlc.arg('order_by') = 'updated_at_asc' then archive.updated_at end asc,
    case when sqlc.arg('order_by') = 'updated_at_desc' then archive.updated_at end desc,
    case
        when sqlc.arg('order_by') = 'release_date_asc' then archive.release_date
    end asc,
    case
        when sqlc.arg('order_by') = 'release_date_desc' then archive.release_date
    end desc,
    case when sqlc.arg('order_by') = 'last_read_asc' then progress.last_read end asc,
    case when sqlc.arg('order_by') = 'last_read_desc' then progress.last_read end desc
limit ?
offset ?
;

-- name: TotalArchiveWithCharacter :one
select count(archive.id)
from archive
join archive_character on archive.id = archive_character.archive_id
join character on archive_character.character_id = character.id
where character.name = ?
;

-- name: DecrementCharacterCount :exec
update character
set count = count - 1
where id in (sqlc.slice('characters'))
;

-- name: IncrementCharacterCount :exec
update character
set count = count + 1
where id in (sqlc.slice('characters'))
;

-- name: DeleteCharacter :exec
delete from character
where id = ?
;

-- name: DeleteAllCharacter :exec
delete from character
;

-- name: BulkAddArchiveCharacters :exec
insert into archive_character (archive_id, character_id)
select ?, value 
from json_each(sqlc.arg('characters'))
;

-- name: BulkAddCharacters :exec
insert or ignore into character (name, count)
select value, 0
from json_each(sqlc.arg('characters'))
;

-- name: BulkGetCharacters :many
select id, name
from character
where name in (sqlc.slice('characters'))
;

-- name: BulkRemoveCharactersFromArchive :exec
delete from archive_character
where
    archive_id = ?
    and character_id in (select value from json_each(sqlc.arg('characters')))
;

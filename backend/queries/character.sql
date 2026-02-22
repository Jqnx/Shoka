-- name: CreateCharacter :one
insert into character (name, count)
values ($1, $2)
returning *
;

-- name: AddCharacterToArchive :exec
insert into archive_character (archive_id, character_id)
values ($1, $2)
;

-- name: BulkAddArchiveCharacters :exec
insert into archive_character (archive_id, character_id)
select $1, unnest(sqlc.arg('characters')::int[])
;

-- name: EnsureCharacterExist :many
insert into character (name, count)
select unnest(sqlc.arg('characters')::text[]), 0
on conflict (name) do update
set name = excluded.name
returning id, name
;

-- name: GetCharacter :one
select *
from character
where name = $1
;

-- name: GetAllCharacter :many
select *
from character
order by id
;

-- name: CharacterExists :execresult
select id, name
from character
where name = $1
;

-- name: RemoveCharacterFromArchive :exec
delete from archive_character
where archive_id = $1 and character_id = any(sqlc.arg('characters')::int[])
;

-- name: GetArchiveCharacters :many
select character.*
from archive
join archive_character on archive.id = archive_character.archive_id
join character on archive_character.character_id = character.id
where archive.id = $1
;

-- name: GetArchiveCharacterIDs :many
select character.id
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

-- name: GetArchiveIDsByCharacters :many
select archive.id
from archive
join archive_character on archive.id = archive_character.archive_id
join character on archive_character.character_id = character.id
where character.name = any(sqlc.arg('characters')::text[])
group by archive.id
having count(distinct character.id) = sqlc.arg('amount')
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
order by
    case when @order_by::text = 'title_asc' then archive.title end asc,
    case when @order_by = 'title_desc' then archive.title end desc nulls last,
    case when @order_by = 'page_count_asc' then archive.page_count end asc,
    case when @order_by = 'page_count_desc' then archive.page_count end desc nulls last,
    case when @order_by = 'created_at_asc' then archive.created_at end asc,
    case when @order_by = 'created_at_desc' then archive.created_at end desc nulls last,
    case when @order_by = 'updated_at_asc' then archive.updated_at end asc,
    case when @order_by = 'updated_at_desc' then archive.updated_at end desc nulls last,
    case when @order_by = 'release_date_asc' then archive.release_date end asc,
    case
        when @order_by = 'release_date_desc' then archive.release_date
    end desc nulls last,
    case when @order_by = 'last_read_asc' then reading_progress.last_read end asc,
    case
        when @order_by = 'last_read_desc' then reading_progress.last_read
    end desc nulls last
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

-- name: DecrementCharacterCount :exec
update character
set count = count - 1
where id = any(sqlc.arg('characters')::int[])
;

-- name: IncrementCharacterCount :exec
update character
set count = count + 1
where id = any(sqlc.arg('characters')::int[])
;

-- name: DeleteCharacter :exec
delete from character
where id = $1
;

-- name: DeleteAllCharacter :exec
delete from character
;


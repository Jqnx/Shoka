-- name: CreateParody :one
insert into parodies (parody, count)
values ($1, $2)
returning *
;

-- name: AddParodyToArchive :exec
insert into archives_parodies (archive_id, parody_id)
values ($1, $2)
;

-- name: GetParody :one
select id, parody, count
from parodies
where parody = $1
;

-- name: GetAllParodies :many
select id, parody, count
from parodies
order by id
;

-- name: ParodyExists :execresult
select id, parody
from parodies
where parody = $1
;

-- name: RemoveParodyFromArchive :many
delete from archives_parodies
where archive_id = $1
returning
    parody_id,
    (
        select parodies.count
        from parodies
        where parodies.id = archives_parodies.parody_id
    )
;

-- name: GetArchiveParodies :many
select parodies.id, parodies.parody, parodies.count
from archives
join archives_parodies on archives.id = archives_parodies.archive_id
join parodies on archives_parodies.parody_id = parodies.id
where archives.archive_id = $1
;

-- name: GetArchivesByParody :many
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
    archives.release_date
from archives
join archives_parodies on archives.id = archives_parodies.archive_id
join parodies on archives_parodies.parody_id = parodies.id
where parodies.parody = $1
;

-- name: GetArchiveIDsByParody :many
select archives.archive_id
from archives
join archives_parodies on archives.id = archives_parodies.archive_id
join parodies on archives_parodies.parody_id = parodies.id
where parodies.parody = $1
;

-- name: GetArchivesByParodyList :many
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
    archives.release_date
from archives
join archives_parodies on archives.id = archives_parodies.archive_id
join parodies on archives_parodies.parody_id = parodies.id
where parodies.parody = $1
limit $2
offset $3
;

-- name: TotalArchivesWithParody :one
select count(archives.archive_id)
from archives
join archives_parodies on archives.id = archives_parodies.archive_id
join parodies on archives_parodies.parody_id = parodies.id
where parodies.parody = $1
;

-- name: UpdateParodyCount :exec
update parodies
set count = $1
where id = $2
;


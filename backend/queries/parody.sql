-- name: CreateParody :one
insert into parody (name, count)
values ($1, $2)
returning *
;

-- name: AddParodyToArchive :exec
insert into archive_parody (archive_id, parody_id)
values ($1, $2)
;

-- name: GetParody :one
select id, name, count
from parody
where name = $1
;

-- name: GetAllParody :many
select id, name, count
from parody
order by id
;

-- name: ParodyExists :execresult
select id, name
from parody
where name = $1
;

-- name: RemoveParodyFromArchive :many
delete from archive_parody
where archive_id = $1
returning
    parody_id,
    (select parody.count from parody where parody.id = archive_parody.parody_id)
;

-- name: GetArchiveParody :many
select parody.id, parody.name, parody.count
from archive
join archive_parody on archive.id = archive_parody.archive_id
join parody on archive_parody.parody_id = parody.id
where archive.id = $1
;

-- name: GetArchiveByParody :many
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
join archive_parody on archive.id = archive_parody.archive_id
join parody on archive_parody.parody_id = parody.id
left join
    reading_progress
    on archive.id = reading_progress.archive_id
    and reading_progress.user_id = sqlc.arg('uid')
where parody.name = $1
;

-- name: GetArchiveIDsByParody :many
select archive.id
from archive
join archive_parody on archive.id = archive_parody.archive_id
join parody on archive_parody.parody_id = parody.id
where parody.name = $1
;

-- name: GetArchiveByParodyList :many
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
join archive_parody on archive.id = archive_parody.archive_id
join parody on archive_parody.parody_id = parody.id
left join
    reading_progress
    on archive.id = reading_progress.archive_id
    and reading_progress.user_id = sqlc.arg('uid')
where parody.name = $1
limit $2
offset $3
;

-- name: TotalArchiveWithParody :one
select count(archive.id)
from archive
join archive_parody on archive.id = archive_parody.archive_id
join parody on archive_parody.parody_id = parody.id
where parody.name = $1
;

-- name: UpdateParodyCount :exec
update parody
set count = $1
where id = $2
;

-- name: DeleteParody :exec
delete from parody
where id = $1
;

-- name: DeleteAllParody :exec
delete from parody
;


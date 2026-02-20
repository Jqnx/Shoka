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
select *
from parody
where name = $1
;

-- name: GetAllParody :many
select *
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

-- name: GetArchiveIDsByParodies :many
select archive.id
from archive
join archive_parody on archive.id = archive_parody.archive_id
join parody on archive_parody.parody_id = parody.id
where parody.name = any(sqlc.arg('parodies')::text[])
group by archive.id
having count(distinct parody.id) = sqlc.arg('amount')
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


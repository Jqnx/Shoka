-- name: CreateParody :one
insert into parody (name, count)
values (?, ?)
returning *
;

-- name: AddParodyToArchive :exec
insert into archive_parody (archive_id, parody_id)
values (?, ?)
;


-- name: EnsureParodyExist :many
insert into parody (name, count)
select value, 0 from json_each(sqlc.arg('parodies'))
on conflict (name) do update
set name = excluded.name
returning id, name
;

-- name: GetParody :one
select *
from parody
where name = ?
;

-- name: GetAllParody :many
select *
from parody
order by id
;

-- name: ParodyExists :execresult
select id, name
from parody
where name = ?
;

-- name: RemoveParodyFromArchive :exec
delete from archive_parody
where archive_id = ? and parody_id in (sqlc.slice('parodies'))
;

-- name: GetArchiveParody :many
select parody.*
from archive
join archive_parody on archive.id = archive_parody.archive_id
join parody on archive_parody.parody_id = parody.id
where archive.id = ?
;

-- name: GetArchiveParodyIDs :many
select parody.id
from archive
join archive_parody on archive.id = archive_parody.archive_id
join parody on archive_parody.parody_id = parody.id
where archive.id = ?
;

-- name: GetArchiveByParody :many
select archive.*, progress.page
from archive
join archive_parody on archive.id = archive_parody.archive_id
join parody on archive_parody.parody_id = parody.id
left join
    progress on archive.id = progress.archive_id and progress.user_id = sqlc.arg('uid')
where parody.name = ?
;

-- name: GetArchiveIDsByParody :many
select archive.id
from archive
join archive_parody on archive.id = archive_parody.archive_id
join parody on archive_parody.parody_id = parody.id
where parody.name = ?
;

-- name: GetArchiveIDsByParodies :many
select archive.id
from archive
join archive_parody on archive.id = archive_parody.archive_id
join parody on archive_parody.parody_id = parody.id
where parody.name in (sqlc.slice('parodies'))
group by archive.id
having count(distinct parody.id) = sqlc.arg('amount')
;

-- name: GetArchiveByParodyList :many
select archive.*, progress.page
from archive
join archive_parody on archive.id = archive_parody.archive_id
join parody on archive_parody.parody_id = parody.id
left join
    progress on archive.id = progress.archive_id and progress.user_id = sqlc.arg('uid')
where parody.name = ?
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

-- name: TotalArchiveWithParody :one
select count(archive.id)
from archive
join archive_parody on archive.id = archive_parody.archive_id
join parody on archive_parody.parody_id = parody.id
where parody.name = ?
;

-- name: DecrementParodyCount :exec
update parody
set count = count - 1
where id in (sqlc.slice('parodies'))
;

-- name: IncrementParodyCount :exec
update parody
set count = count + 1
where id in (sqlc.slice('parodies'))
;

-- name: DeleteParody :exec
delete from parody
where id = ?
;

-- name: DeleteAllParody :exec
delete from parody
;

-- name: BulkAddArchiveParodies :exec
insert into archive_parody (archive_id, parody_id)
select ?, value 
from json_each(sqlc.arg('parodies'))
;

-- name: BulkAddParodies :exec
insert or ignore into parody (name, count)
select value, 0
from json_each(sqlc.arg('parodies'))
;

-- name: BulkGetParodies :many
select id, name
from parody
where name in (sqlc.slice('parodies'))
;

-- name: BulkRemoveParodiesFromArchive :exec
delete from archive_parody
where
    archive_id = ? and parody_id in (select value from json_each(sqlc.arg('parodies')))
;

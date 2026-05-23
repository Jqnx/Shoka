-- name: CreateTag :one
insert into tag (name, count)
values (?, ?)
returning *
;

-- name: AddTagToArchive :exec
insert into archive_tag (archive_id, tag_id)
values (?, ?)
;

-- name: BulkAddArchiveTags :exec
insert or ignore into archive_tag (archive_id, tag_id)
select ?, value
from json_each(sqlc.arg('tags'))
;

-- name: BulkAddTags :exec
insert or ignore into tag (name, count)
select value, 0
from json_each(sqlc.arg('tags'))
;

-- name: BulkGetTags :many
select id, name
from tag
where name in (sqlc.slice('tags'))
;

-- name: EnsureTagExist :many
insert into tag (name, count)
select value, 0 from json_each(sqlc.arg('tags'))
on conflict (name) do update
set name = excluded.name
returning id, name
;

-- name: GetTag :one
select *
from tag
where name = ?
;

-- name: GetAllTags :many
select *
from tag
order by name
;

-- name: GetArchivesByTagName :many
select archive.*, progress.page, progress.last_read, progress.completed
from archive
join archive_tag on archive.id = archive_tag.archive_id
join tag on archive_tag.tag_id = tag.id
left join progress on archive.id = progress.archive_id and progress.user_id = sqlc.arg('uid')
where tag.name = sqlc.arg('name')
order by archive.title asc
limit sqlc.arg('limit')
offset sqlc.arg('offset')
;

-- name: CountTags :one
select count(*) from tag
;

-- name: GetTagList :many
select *
from tag
order by name
limit ?
offset ?
;

-- name: TagExists :execresult
select id, name
from tag
where name = ?
;

-- name: RemoveTagFromArchive :exec
delete from archive_tag
where archive_id = ? and tag_id in (sqlc.slice('tags'))
;

-- name: BulkRemoveTagFromArchive :exec
delete from archive_tag
where archive_id = ? and tag_id in (select value from json_each(sqlc.arg('tags')))
;

-- name: GetArchiveTag :many
select tag.*
from archive
join archive_tag on archive.id = archive_tag.archive_id
join tag on archive_tag.tag_id = tag.id
where archive.id = ?
;

-- name: GetArchiveTagIDs :many
select tag.id
from archive
join archive_tag on archive.id = archive_tag.archive_id
join tag on archive_tag.tag_id = tag.id
where archive.id = ?
;

-- name: GetArchiveByTag :many
select archive.*, progress.page
from archive
join archive_tag on archive.id = archive_tag.archive_id
join tag on archive_tag.tag_id = tag.id
left join
    progress on archive.id = progress.archive_id and progress.user_id = sqlc.arg('uid')
where tag.name = ?
;

-- name: GetArchiveIDsByTag :many
select archive.id
from archive
join archive_tag on archive.id = archive_tag.archive_id
join tag on archive_tag.tag_id = tag.id
where tag.name = ?
;

-- name: GetArchiveIDsByTags :many
select archive.id
from archive
join archive_tag on archive.id = archive_tag.archive_id
join tag on archive_tag.tag_id = tag.id
where tag.name in (sqlc.slice('tags'))
group by archive.id
having count(distinct tag.id) = sqlc.arg('amount')
;

-- name: GetArchiveByTagList :many
select archive.*, progress.page
from archive
join archive_tag on archive.id = archive_tag.archive_id
join tag on archive_tag.tag_id = tag.id
left join
    progress on archive.id = progress.archive_id and progress.user_id = sqlc.arg('uid')
where tag.name = ?
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

-- name: TotalArchiveWithTag :one
select count(archive.id)
from archive
join archive_tag on archive.id = archive_tag.archive_id
join tag on archive_tag.tag_id = tag.id
where tag.name = ?
;

-- name: DecrementTagCount :exec
update tag
set count = count - 1
where id in (sqlc.slice('tags'))
;

-- name: IncrementTagCount :exec
update tag
set count = count + 1
where id in (sqlc.slice('tags'))
;

-- name: UpdateTagDescription :exec
update tag
set description = ?
where id = ?
;

-- name: DeleteTag :exec
delete from tag
where id = ?
;

-- name: DeleteAllTag :exec
delete from tag
;

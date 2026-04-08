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
insert into archive_tag (archive_id, tag_id)
select ?, value from json_each(sqlc.arg('tags'))
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

-- name: TagExists :execresult
select id, name
from tag
where name = ?
;

-- name: RemoveTagFromArchive :exec
delete from archive_tag
where archive_id = ? and tag_id in (sqlc.slice('tags'))
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
select archive.*, reading_progress.page
from archive
join archive_tag on archive.id = archive_tag.archive_id
join tag on archive_tag.tag_id = tag.id
left join
    reading_progress
    on archive.id = reading_progress.archive_id
    and reading_progress.user_id = sqlc.arg('uid')
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
select archive.*, reading_progress.page
from archive
join archive_tag on archive.id = archive_tag.archive_id
join tag on archive_tag.tag_id = tag.id
left join
    reading_progress
    on archive.id = reading_progress.archive_id
    and reading_progress.user_id = sqlc.arg('uid')
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
    case
        when sqlc.arg('order_by') = 'last_read_asc' then reading_progress.last_read
    end asc,
    case
        when sqlc.arg('order_by') = 'last_read_desc' then reading_progress.last_read
    end desc
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

-- name: DeleteTag :exec
delete from tag
where id = ?
;

-- name: DeleteAllTag :exec
delete from tag
;

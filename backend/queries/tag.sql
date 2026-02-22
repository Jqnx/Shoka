-- name: CreateTag :one
insert into tag (name, count)
values ($1, $2)
returning *
;

-- name: AddTagToArchive :exec
insert into archive_tag (archive_id, tag_id)
values ($1, $2)
;

-- name: BulkAddArchiveTags :exec
insert into archive_tag (archive_id, tag_id)
select $1, unnest(sqlc.arg('tags')::int[])
;

-- name: EnsureTagExist :many
insert into tag (name, count)
select unnest(sqlc.arg('tags')::text[]), 0
on conflict (name) do update
set name = excluded.name
returning id, name
;

-- name: GetTag :one
select *
from tag
where name = $1
;

-- name: GetAllTags :many
select *
from tag
order by name
;

-- name: TagExists :execresult
select id, name
from tag
where name = $1
;

-- name: RemoveTagFromArchive :exec
delete from archive_tag
where archive_id = $1 and tag_id = any(sqlc.arg('tags')::int[])
;

-- name: GetArchiveTag :many
select tag.*
from archive
join archive_tag on archive.id = archive_tag.archive_id
join tag on archive_tag.tag_id = tag.id
where archive.id = $1
;

-- name: GetArchiveTagIDs :many
select tag.id
from archive
join archive_tag on archive.id = archive_tag.archive_id
join tag on archive_tag.tag_id = tag.id
where archive.id = $1
;

-- name: GetArchiveByTag :many
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
join archive_tag on archive.id = archive_tag.archive_id
join tag on archive_tag.tag_id = tag.id
left join
    reading_progress
    on archive.id = reading_progress.archive_id
    and reading_progress.user_id = sqlc.arg('uid')
where tag.name = $1
;

-- name: GetArchiveIDsByTag :many
select archive.id
from archive
join archive_tag on archive.id = archive_tag.archive_id
join tag on archive_tag.tag_id = tag.id
where tag.name = $1
;

-- name: GetArchiveIDsByTags :many
select archive.id
from archive
join archive_tag on archive.id = archive_tag.archive_id
join tag on archive_tag.tag_id = tag.id
where tag.name = any(sqlc.arg('tags')::text[])
group by archive.id
having count(distinct tag.id) = sqlc.arg('amount')
;

-- name: GetArchiveByTagList :many
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
join archive_tag on archive.id = archive_tag.archive_id
join tag on archive_tag.tag_id = tag.id
left join
    reading_progress
    on archive.id = reading_progress.archive_id
    and reading_progress.user_id = sqlc.arg('UserID')
where tag.name = $1
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

-- name: TotalArchiveWithTag :one
select count(archive.id)
from archive
join archive_tag on archive.id = archive_tag.archive_id
join tag on archive_tag.tag_id = tag.id
where tag.name = $1
;

-- name: DecrementTagCount :exec
update tag
set count = count - 1
where id = any(sqlc.arg('tags')::int[])
;

-- name: IncrementTagCount :exec
update tag
set count = count + 1
where id = any(sqlc.arg('tags')::int[])
;

-- name: DeleteTag :exec
delete from tag
where id = $1
;

-- name: DeleteAllTag :exec
delete from tag
;


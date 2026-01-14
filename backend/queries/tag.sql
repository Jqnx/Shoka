-- name: CreateTag :one
insert into tag (name, count)
values ($1, $2)
returning *
;

-- name: AddTagToArchive :exec
insert into archive_tag (archive_id, tag_id)
values ($1, $2)
;

-- name: GetTag :one
select id, name, count
from tag
where name = $1
;

-- name: GetAllTags :many
select id, name, count
from tag
order by name
;

-- name: TagExists :execresult
select id, name
from tag
where name = $1
;

-- name: RemoveTagFromArchive :many
delete from archive_tag
where archive_id = $1
returning tag_id, (select tag.count from tag where tag.id = archive_tag.tag_id)
;

-- name: GetArchiveTag :many
select tag.id, tag.name, tag.count
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
    and reading_progress.user_id = sqlc.arg('uid')
where tag.name = $1
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

-- name: UpdateTagCount :exec
update tag
set count = $1
where id = $2
;

-- name: DeleteTag :exec
delete from tag
where id = $1
;

-- name: DeleteAllTag :exec
delete from tag
;


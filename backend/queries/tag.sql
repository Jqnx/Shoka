-- name: CreateTag :one
insert into tags (name, count)
values ($1, $2)
returning *
;

-- name: AddTagToArchive :exec
insert into archives_tags (archive_id, tag_id)
values ($1, $2)
;

-- name: GetTag :one
select id, name, count
from tags
where name = $1
;

-- name: GetAllTags :many
select id, name, count
from tags
order by name
;

-- name: TagExists :execresult
select id, name
from tags
where name = $1
;

-- name: RemoveTagFromArchive :many
delete from archives_tags
where archive_id = $1
returning tag_id, (select tags.count from tags where tags.id = archives_tags.tag_id)
;

-- name: GetArchiveTags :many
select tags.id, tags.name, tags.count
from archives
join archives_tags on archives.id = archives_tags.archive_id
join tags on archives_tags.tag_id = tags.id
where archives.id = $1
;

-- name: GetArchivesByTag :many
select
    archives.id,
    archives.title,
    archives.summary,
    archives.language,
    archives.category,
    archives.page_count,
    archives.file_path,
    archives.file_name,
    archives.hash,
    archives.thumbs_path,
    archives.cover_path,
    archives.cover_img,
    archives.type,
    archives.created_at,
    archives.updated_at,
    archives.release_date,
    reading_progress.page
from archives
join archives_tags on archives.id = archives_tags.archive_id
join tags on archives_tags.tag_id = tags.id
left join
    reading_progress
    on archives.id = reading_progress.archive_id
    and reading_progress.user_id = sqlc.arg('uid')
where tags.name = $1
;


-- name: GetArchiveIDsByTag :many
select archives.id
from archives
join archives_tags on archives.id = archives_tags.archive_id
join tags on archives_tags.tag_id = tags.id
where tags.name = $1
;

-- name: GetArchivesByTagList :many
select
    archives.id,
    archives.title,
    archives.summary,
    archives.language,
    archives.category,
    archives.page_count,
    archives.file_path,
    archives.file_name,
    archives.hash,
    archives.thumbs_path,
    archives.cover_path,
    archives.cover_img,
    archives.type,
    archives.created_at,
    archives.updated_at,
    archives.release_date,
    reading_progress.page
from archives
join archives_tags on archives.id = archives_tags.archive_id
join tags on archives_tags.tag_id = tags.id
left join
    reading_progress
    on archives.id = reading_progress.archive_id
    and reading_progress.user_id = sqlc.arg('uid')
where tags.name = $1
limit $2
offset $3
;


-- name: TotalArchivesWithTag :one
select count(archives.id)
from archives
join archives_tags on archives.id = archives_tags.archive_id
join tags on archives_tags.tag_id = tags.id
where tags.name = $1
;

-- name: UpdateTagCount :exec
update tags
set count = $1
where id = $2
;


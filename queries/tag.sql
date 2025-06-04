-- name: CreateTag :one
insert into tags (tag, count)
values ($1, $2)
returning *
;

-- name: AddTagToArchive :exec
insert into archives_tags (archive_id, tag_id)
values ($1, $2)
;

-- name: GetTag :one
select id, tag, count
from tags
where tag = $1
;

-- name: GetAllTags :many
select id, tag, count
from tags
order by tag
;

-- name: TagExists :execresult
select id, tag
from tags
where tag = $1
;

-- name: RemoveTagFromArchive :many
delete from archives_tags
where archive_id = $1
returning tag_id, (select tags.count from tags where tags.id = archives_tags.tag_id)
;

-- name: GetArchiveTags :many
select tags.id, tags.tag, tags.count
from archives
join archives_tags on archives.id = archives_tags.archive_id
join tags on archives_tags.tag_id = tags.id
where archives.archive_id = $1
;

-- name: GetArchivesByTag :many
select archives.*
from archives
join archives_tags on archives.id = archives_tags.archive_id
join tags on archives_tags.tag_id = tags.id
where tags.tag = $1
;

-- name: GetArchivesByTagList :many
select archives.*
from archives
join archives_tags on archives.id = archives_tags.archive_id
join tags on archives_tags.tag_id = tags.id
where tags.tag = $1
limit $2
offset $3
;

-- name: TotalArchivesWithTag :one
select count(archives.archive_id)
from archives
join archives_tags on archives.id = archives_tags.archive_id
join tags on archives_tags.tag_id = tags.id
where tags.tag = $1
;

-- name: UpdateTagCount :exec
update tags
set count = $1
where id = $2
;


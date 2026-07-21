-- name: GetArchiveFilterSortList :many
select archive.*, reading_progress.page, reading_progress.last_read, reading_progress.completed
from archive
left join
    reading_progress on archive.id = reading_progress.archive_id and reading_progress.user_id = sqlc.arg('uid')
where archive.id in (sqlc.slice('ids'))
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
    case when sqlc.arg('order_by') = 'last_read_asc' then reading_progress.last_read end asc,
    case when sqlc.arg('order_by') = 'last_read_desc' then reading_progress.last_read end desc
limit ?
offset ?
;

-- name: GetArchiveFilter :one
select archive.id
from archive
where archive.id in (sqlc.slice('ids'))
limit ?
offset ?
;

-- name: GetArchiveSortList :many
select archive.*, reading_progress.page, reading_progress.last_read, reading_progress.completed
from archive
left join
    reading_progress on archive.id = reading_progress.archive_id and reading_progress.user_id = sqlc.arg('uid')
order by
    case when sqlc.arg('order_by') = 'title_asc' then title end asc,
    case when sqlc.arg('order_by') = 'title_desc' then title end desc,
    case when sqlc.arg('order_by') = 'page_count_asc' then page_count end asc,
    case when sqlc.arg('order_by') = 'page_count_desc' then page_count end desc,
    case when sqlc.arg('order_by') = 'created_at_asc' then created_at end asc,
    case when sqlc.arg('order_by') = 'created_at_desc' then created_at end desc,
    case when sqlc.arg('order_by') = 'updated_at_asc' then updated_at end asc,
    case when sqlc.arg('order_by') = 'updated_at_desc' then updated_at end desc,
    case when sqlc.arg('order_by') = 'release_date_asc' then release_date end asc,
    case when sqlc.arg('order_by') = 'release_date_desc' then release_date end desc,
    case when sqlc.arg('order_by') = 'last_read_asc' then reading_progress.last_read end asc,
    case when sqlc.arg('order_by') = 'last_read_desc' then reading_progress.last_read end desc
limit ?
offset ?
;

-- name: GetArchiveSort :many
select archive.*, reading_progress.page, reading_progress.last_read, reading_progress.completed
from archive
left join
    reading_progress on archive.id = reading_progress.archive_id and reading_progress.user_id = sqlc.arg('uid')
order by
    case when sqlc.arg('order_by') = 'title_asc' then title end asc,
    case when sqlc.arg('order_by') = 'title_desc' then title end desc,
    case when sqlc.arg('order_by') = 'page_count_asc' then page_count end asc,
    case when sqlc.arg('order_by') = 'page_count_desc' then page_count end desc,
    case when sqlc.arg('order_by') = 'created_at_asc' then created_at end asc,
    case when sqlc.arg('order_by') = 'created_at_desc' then created_at end desc,
    case when sqlc.arg('order_by') = 'updated_at_asc' then updated_at end asc,
    case when sqlc.arg('order_by') = 'updated_at_desc' then updated_at end desc,
    case when sqlc.arg('order_by') = 'release_date_asc' then release_date end asc,
    case when sqlc.arg('order_by') = 'release_date_desc' then release_date end desc,
    case when sqlc.arg('order_by') = 'last_read_asc' then reading_progress.last_read end asc,
    case when sqlc.arg('order_by') = 'last_read_desc' then reading_progress.last_read end desc
;

-- name: CountFilteredArchive :one
select count(archive.id)
from archive
where archive.id in (sqlc.slice('ids'))
;

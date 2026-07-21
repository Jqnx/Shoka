-- name: GetAllCategory :many
select distinct category
from archive
where category is not null
order by category
;

-- name: GetArchiveByCategory :many
select archive.*, reading_progress.page
from archive
left join
    reading_progress on archive.id = reading_progress.archive_id and reading_progress.user_id = sqlc.arg('uid')
where category = ?
;

-- name: GetArchiveIDsByCategory :many
select archive.id
from archive
where category = ?
;

-- name: GetArchiveIDsByCategories :many
select archive.id
from archive
where category in (sqlc.slice('categories'))
group by archive.id
having count(distinct archive.id) = sqlc.arg('amount')
;

-- name: GetArchiveByCategoryList :many
select archive.*, reading_progress.page
from archive
left join
    reading_progress on archive.id = reading_progress.archive_id and reading_progress.user_id = sqlc.arg('uid')
where category = ?
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
        when sqlc.arg('order_by') = 'release_date_desc' then archive.relase_date
    end desc,
    case when sqlc.arg('order_by') = 'last_read_asc' then reading_progress.last_read end asc,
    case when sqlc.arg('order_by') = 'last_read_desc' then reading_progress.last_read end desc
limit ?
offset ?
;

-- name: TotalArchiveWithCategory :one
select count(id)
from archive
where category = ?
;

-- name: CategoryExists :execresult
select category
from archive
where category = ?
;

-- name: GetArchivesFilterSortList :many
select
    archives.id,
    archives.title,
    archives.summary,
    archives.language,
    archives.category,
    archives.page_count,
    archives.file_path,
    archives.file_name,
    -- archives.archive_id,
    archives.hash,
    archives.thumbs_path,
    archives.cover_path,
    archives.type,
    archives.created_at,
    archives.updated_at,
    archives.release_date,
    reading_progress.page,
    reading_progress.last_read,
    reading_progress.state
from archives
left join
    reading_progress
    on archives.id = reading_progress.archive_id
    and reading_progress.user_id = sqlc.arg('uid')
where archives.id = any(sqlc.arg('ids')::text[])
order by
    case when @order_by::text = 'title_asc' then archives.title end asc,
    case when @order_by = 'title_desc' then archives.title end desc nulls last,
    case when @order_by = 'page_count_asc' then archives.page_count end asc,
    case
        when @order_by = 'page_count_desc' then archives.page_count
    end desc nulls last,
    case when @order_by = 'created_at_asc' then archives.created_at end asc,
    case
        when @order_by = 'created_at_desc' then archives.created_at
    end desc nulls last,
    case when @order_by = 'updated_at_asc' then archives.updated_at end asc,
    case
        when @order_by = 'updated_at_desc' then archives.updated_at
    end desc nulls last,
    case when @order_by = 'release_date_asc' then archives.release_date end asc,
    case
        when @order_by = 'release_date_desc' then archives.release_date
    end desc nulls last,
    case when @order_by = 'last_read_asc' then reading_progress.last_read end asc,
    case
        when @order_by = 'last_read_desc' then reading_progress.last_read
    end desc nulls last
limit $1
offset $2
;

-- name: GetArchivesFilter :one
select archives.id
from archives
where archives.id = any(sqlc.arg('ids')::text[])
limit $1
offset $2
;

-- name: GetArchiveSortList :many
select
    archives.id,
    archives.title,
    archives.summary,
    archives.language,
    archives.category,
    archives.page_count,
    archives.file_path,
    archives.file_name,
    -- archives.archive_id,
    archives.hash,
    archives.thumbs_path,
    archives.cover_path,
    archives.type,
    archives.created_at,
    archives.updated_at,
    archives.release_date,
    reading_progress.page,
    reading_progress.last_read,
    reading_progress.state
from archives
left join
    reading_progress
    on archives.id = reading_progress.archive_id
    and reading_progress.user_id = sqlc.arg('uid')
order by
    case when @order_by::text = 'title_asc' then title end asc,
    case when @order_by = 'title_desc' then title end desc nulls last,
    case when @order_by = 'page_count_asc' then page_count end asc,
    case when @order_by = 'page_count_desc' then page_count end desc nulls last,
    case when @order_by = 'created_at_asc' then created_at end asc,
    case when @order_by = 'created_at_desc' then created_at end desc nulls last,
    case when @order_by = 'updated_at_asc' then updated_at end asc,
    case when @order_by = 'updated_at_desc' then updated_at end desc nulls last,
    case when @order_by = 'release_date_asc' then release_date end asc,
    case when @order_by = 'release_date_desc' then release_date end desc nulls last,
    case when @order_by = 'last_read_asc' then reading_progress.last_read end asc,
    case
        when @order_by = 'last_read_desc' then reading_progress.last_read
    end desc nulls last
limit $1
offset $2
;

-- name: GetArchiveSort :many
select
    archives.id,
    archives.title,
    archives.summary,
    archives.language,
    archives.category,
    archives.page_count,
    archives.file_path,
    archives.file_name,
    -- archives.archive_id,
    archives.hash,
    archives.thumbs_path,
    archives.cover_path,
    archives.type,
    archives.created_at,
    archives.updated_at,
    archives.release_date,
    reading_progress.page,
    reading_progress.last_read,
    reading_progress.state
from archives
left join
    reading_progress
    on archives.id = reading_progress.archive_id
    and reading_progress.user_id = sqlc.arg('uid')
order by
    case when @order_by::text = 'title_asc' then title end asc,
    case when @order_by = 'title_desc' then title end desc nulls last,
    case when @order_by = 'page_count_asc' then page_count end asc,
    case when @order_by = 'page_count_desc' then page_count end desc nulls last,
    case when @order_by = 'created_at_asc' then created_at end asc,
    case when @order_by = 'created_at_desc' then created_at end desc nulls last,
    case when @order_by = 'updated_at_asc' then updated_at end asc,
    case when @order_by = 'updated_at_desc' then updated_at end desc nulls last,
    case when @order_by = 'release_date_asc' then release_date end asc,
    case when @order_by = 'release_date_desc' then release_date end desc nulls last,
    case when @order_by = 'last_read_asc' then reading_progress.last_read end asc,
    case
        when @order_by = 'last_read_desc' then reading_progress.last_read
    end desc nulls last
;

-- name: CountFilteredArchives :one
select count(archives.id)
from archives
where archives.id = any(sqlc.arg('ids')::text[])
;


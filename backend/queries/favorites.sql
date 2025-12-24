-- name: AddFavoriteArchive :exec
insert into favorite_archives (archive_id, user_id, favorited_at)
values ($1, $2, $3)
;

-- name: GetUserFavoriteArchiveList :many
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
    archive.release_date
from archive
join favorite_archives on archive.id = favorite_archives.archive_id
join "user" as u on favorite_archives.user_id = u.id
left join
    reading_progress
    on archive.id = reading_progress.archive_id
    and reading_progress.user_id = u.id
where u.id = $1
limit $2
offset $3
;

-- name: GetUserFavoriteArchiveShuffle :one
select archive.id
from archive
join favorite_archives on archive.id = favorite_archives.archive_id
join "user" as u on favorite_archives.user_id = u.id
where u.id = $1
limit $2
offset $3
;

-- name: GetUserFavoriteArchiveAll :many
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
    archive.release_date
from archive
join favorite_archives on archive.id = favorite_archives.archive_id
join "user" as u on favorite_archives.user_id = u.id
left join
    reading_progress
    on archive.id = reading_progress.archive_id
    and reading_progress.user_id = u.id
where u.id = $1
order by favorited_at desc
;

-- name: GetFavoriteArchiveFilterSortList :many
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
join favorite_archives on archive.id = favorite_archives.archive_id
left join
    reading_progress
    on archive.id = reading_progress.archive_id
    and reading_progress.user_id = sqlc.arg(user_id)::uuid
where
    archive.id = any(sqlc.arg('ids')::text[])
    and archive.id in (
        select archive.id
        from archive
        join favorite_archives on archive.id = favorite_archives.archive_id
        join "user" as u on favorite_archives.user_id = u.id
        where u.id = sqlc.arg(user_id)::uuid
    )
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
    case
        when @order_by = 'favorited_at_asc' then favorite_archives.favorited_at
    end asc,
    case
        when @order_by = 'favorited_at_desc' then favorite_archives.favorited_at
    end desc nulls last
limit $1
offset $2
;

-- name: GetFavoriteArchiveFilter :one
select archive.id
from archive
join favorite_archives on archive.id = favorite_archives.archive_id
where
    archive.id = any(sqlc.arg('ids')::text[])
    and archive.id in (
        select archive.id
        from archive
        join favorite_archives on archive.id = favorite_archives.archive_id
        join "user" as u on favorite_archives.user_id = u.id
        where u.id = sqlc.arg(user_id)::uuid
    )
limit $1
offset $2
;

-- name: CountFavoriteFilteredArchive :one
select count(archive.id)
from archive
where
    archive.id = any(sqlc.arg('ids')::text[])
    and archive.id in (
        select archive.id
        from archive
        join favorite_archives on archive.id = favorite_archives.archive_id
        join "user" as u on favorite_archives.user_id = u.id
        where u.id = sqlc.arg(user_id)::uuid
    )
;

-- name: GetFavoriteArchiveSortList :many
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
join favorite_archives on archive.id = favorite_archives.archive_id
join "user" as u on favorite_archives.user_id = u.id
left join
    reading_progress
    on archive.id = reading_progress.archive_id
    and reading_progress.user_id = sqlc.arg(user_id)::uuid
where u.id = sqlc.arg(user_id)::uuid
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
    case
        when @order_by = 'favorited_at_asc' then favorite_archives.favorited_at
    end asc,
    case
        when @order_by = 'favorited_at_desc' then favorite_archives.favorited_at
    end desc nulls last
limit $1
offset $2
;

-- name: RemoveFavoriteArchive :exec
delete from favorite_archives
where archive_id = $1 and user_id = $2
;

-- name: ArchiveIsFavorited :execresult
select archive_id
from favorite_archives
where user_id = $1 and archive_id = $2
;

-- name: CountUserFavoriteArchive :one
select count(archive_id)
from favorite_archives
where user_id = $1
;


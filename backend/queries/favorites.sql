-- name: AddFavoriteArchive :exec
insert into favorite_archives (archive_id, user_id, favorited_at)
values ($1, $2, $3)
;

-- name: GetUserFavoriteArchivesList :many
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
    archives.release_date
from archives
join favorite_archives on archives.id = favorite_archives.archive_id
join users on favorite_archives.user_id = users.id
left join
    reading_progress
    on archives.id = reading_progress.archive_id
    and reading_progress.user_id = users.id
where users.id = $1
limit $2
offset $3
;

-- name: GetUserFavoriteArchivesShuffle :one
select archives.id
from archives
join favorite_archives on archives.id = favorite_archives.archive_id
join users on favorite_archives.user_id = users.id
where users.id = $1
limit $2
offset $3
;

-- name: GetUserFavoriteArchivesAll :many
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
    archives.release_date
from archives
join favorite_archives on archives.id = favorite_archives.archive_id
join users on favorite_archives.user_id = users.id
left join
    reading_progress
    on archives.id = reading_progress.archive_id
    and reading_progress.user_id = users.id
where users.id = $1
order by favorited_at desc
;

-- name: GetFavoriteArchivesFilterSortList :many
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
join favorite_archives on archives.id = favorite_archives.archive_id
left join
    reading_progress
    on archives.id = reading_progress.archive_id
    and reading_progress.user_id = sqlc.arg(user_id)::uuid
where
    archives.id = any(sqlc.arg('ids')::text[])
    and archives.id in (
        select archives.id
        from archives
        join favorite_archives on archives.id = favorite_archives.archive_id
        join users on favorite_archives.user_id = users.id
        where users.id = sqlc.arg(user_id)::uuid
    )
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
    case
        when @order_by = 'favorited_at_asc' then favorite_archives.favorited_at
    end asc,
    case
        when @order_by = 'favorited_at_desc' then favorite_archives.favorited_at
    end desc nulls last
limit $1
offset $2
;

-- name: GetFavoriteArchivesFilter :one
select archives.id
from archives
join favorite_archives on archives.id = favorite_archives.archive_id
where
    archives.id = any(sqlc.arg('ids')::text[])
    and archives.id in (
        select archives.id
        from archives
        join favorite_archives on archives.id = favorite_archives.archive_id
        join users on favorite_archives.user_id = users.id
        where users.id = sqlc.arg(user_id)::uuid
    )
limit $1
offset $2
;

-- name: CountFavoriteFilteredArchives :one
select count(archives.id)
from archives
where
    archives.id = any(sqlc.arg('ids')::text[])
    and archives.id in (
        select archives.id
        from archives
        join favorite_archives on archives.id = favorite_archives.archive_id
        join users on favorite_archives.user_id = users.id
        where users.id = sqlc.arg(user_id)::uuid
    )
;

-- name: GetFavoriteArchiveSortList :many
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
join favorite_archives on archives.id = favorite_archives.archive_id
join users on favorite_archives.user_id = users.id
left join
    reading_progress
    on archives.id = reading_progress.archive_id
    and reading_progress.user_id = sqlc.arg(user_id)::uuid
where users.id = sqlc.arg(user_id)::uuid
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

-- name: CountUserFavoriteArchives :one
select count(archive_id)
from favorite_archives
where user_id = $1
;


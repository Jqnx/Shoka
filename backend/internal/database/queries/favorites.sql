-- name: AddFavoriteArchive :exec
insert into favorite_archive (archive_id, user_id, favorited_at)
values (?, ?, ?)
;

-- name: GetUserFavoriteArchiveList :many
select archive.*, rp.page
from archive
join favorite_archive as fa on archive.id = fa.archive_id
left join
    reading_progress as rp on archive.id = rp.archive_id and rp.user_id = fa.user_id
where fa.user_id = sqlc.arg('uid')
limit ?
offset ?
;

-- name: GetUserFavoriteArchiveShuffle :one
select archive.id
from archive
join favorite_archive as fa on archive.id = fa.archive_id
where fa.user_id = sqlc.arg('uid')
limit ?
offset ?
;

-- name: GetUserFavoriteArchiveAll :many
select archive.*
from archive
join favorite_archive as fa on archive.id = fa.archive_id
left join
    reading_progress as rp on archive.id = rp.archive_id and rp.user_id = fa.user_id
where fa.user_id = sqlc.arg('uid')
order by fa.favorited_at desc
;

-- name: GetFavoriteArchiveFilterSortList :many
select archive.*, rp.page
from archive
join favorite_archive as fa on archive.id = fa.archive_id
left join
    reading_progress as rp
    on archive.id = rp.archive_id
    and rp.user_id = sqlc.arg('uid')
where
    archive.id in (sqlc.slice('ids'))
    and archive.id in (
        select archive.id
        from archive
        join favorite_archive as fa on archive.id = fa.archive_id
        where fa.user_id = sqlc.arg('uid')
    )
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
    case when sqlc.arg('order_by') = 'favorited_at_asc' then fa.favorited_at end asc,
    case when sqlc.arg('order_by') = 'favorited_at_desc' then fa.favorited_at end desc
limit ?
offset ?
;

-- name: GetFavoriteArchiveFilter :one
select archive.id
from archive
join favorite_archive as fa on archive.id = fa.archive_id
where
    archive.id in (sqlc.slice('ids'))
    and archive.id in (
        select archive.id
        from archive
        join favorite_archive as fa on archive.id = fa.archive_id
        where fa.user_id = sqlc.arg('uid')
    )
limit ?
offset ?
;

-- name: CountFavoriteFilteredArchive :one
select count(archive.id)
from archive
where
    archive.id in (sqlc.slice('ids'))
    and archive.id in (
        select archive.id
        from archive
        join favorite_archive as fa on archive.id = fa.archive_id
        where fa.user_id = sqlc.arg('uid')
    )
;

-- name: GetFavoriteArchiveSortList :many
select archive.*, rp.page
from archive
join favorite_archive as fa on archive.id = fa.archive_id
left join
    reading_progress as rp
    on archive.id = rp.archive_id
    and rp.user_id = sqlc.arg('uid')
where fa.user_id = sqlc.arg('uid')
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
    case when sqlc.arg('order_by') = 'favorited_at_asc' then fa.favorited_at end asc,
    case when sqlc.arg('order_by') = 'favorited_at_desc' then fa.favorited_at end desc
limit ?
offset ?
;

-- name: RemoveFavoriteArchive :exec
delete from favorite_archive
where archive_id = ? and user_id = sqlc.arg('uid')
;

-- name: ArchiveIsFavorited :execresult
select archive_id
from favorite_archive
where user_id = sqlc.arg('uid') and archive_id = ?
;

-- name: CountUserFavoriteArchive :one
select count(archive_id)
from favorite_archive
where user_id = sqlc.arg('uid')
;

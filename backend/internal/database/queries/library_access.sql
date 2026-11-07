-- name: ListLibraryAccessForUser :many
select library_id from library_access where user_id = ?;

-- name: DeleteLibraryAccessForUser :exec
delete from library_access where user_id = ?;

-- name: InsertLibraryAccess :exec
insert into library_access (user_id, library_id) values (?, ?);

/*
-- name: CreateUser :one
insert into users (
  id,
  name,
  password,
  created_at,
  updated_at
) values ( $1, $2, $3, $4, $5 )
returning id, name, password, created_at, updated_at;

-- name: GetUserByID :one
select id, name, password, created_at, updated_at
from users
where id = $1
;

-- name: GetUserByName :one
select id, name, password, created_at, updated_at
from users
where name = $1
;

-- name: UpdateUser :one
update users
set name = coalesce(sqlc.narg('name'), name),
    password = coalesce(sqlc.narg('password'), password),
    updated_at = coalesce($1, updated_at)
where id = $2
returning *
;

-- name: DeleteUser :exec
delete from users
where id = $1
;
*/


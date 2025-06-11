-- name: CreateUser :one
insert into users (
  name,
  password,
  created_at,
  updated_at
) values ( $1, $2, $3, $4 )
returning *;

-- name: GetUserByID :one
select *
from users
where id = $1
;

-- name: GetUserByName :one
select *
from users
where name = $1
;

-- name: GetUserByToken :one
select *
from users
where session = $1
;

-- name: UpdateSession :exec
update users
set session = $1,
    session_expiry = $2
where id = $3
;


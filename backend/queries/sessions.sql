/*
-- name: CreateSession :exec
insert into sessions (
  session_id,
  user_id,
  token,
  expires_at,
  created_at,
  ip_address,
  user_agent
) values ( $1, $2, $3, $4, $5,  $6, $7 )
;

-- name: GetSession :one
select *
from sessions
where token = $1
;

-- name: GetUserByToken :one
select users.id, users.name, sessions.expires_at
from users
join sessions on users.id = sessions.user_id
where sessions.token = $1
;

-- name: GetUserSessions :many
select *
from sessions
where user_id = $1
;

-- name: DeleteSession :exec
delete from sessions
where token = $1
;
*/


-- name: BulkAddVoiceActors :exec
insert or ignore into voice_actor (name, count)
select value, 0
from json_each(sqlc.arg('voice_actors'))
;

-- name: BulkGetVoiceActors :many
select id, name
from voice_actor
where name in (sqlc.slice('voice_actors'))
;

-- name: GetAllVoiceActors :many
select *
from voice_actor
order by name
;

-- name: GetVoiceActorList :many
select *
from voice_actor
order by name
limit ?
offset ?
;

-- name: TotalVoiceActors :one
select count(*)
from voice_actor
;

-- name: DeleteVoiceActor :exec
delete from voice_actor
where id = ?
;

-- name: IncrementVoiceActorCount :exec
update voice_actor
set count = count + 1
where id in (sqlc.slice('voice_actors'))
;

-- name: DecrementVoiceActorCount :exec
update voice_actor
set count = count - 1
where id in (sqlc.slice('voice_actors'))
;

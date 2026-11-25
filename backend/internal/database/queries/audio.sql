-- name: CreateAudio :one
insert into audio (id, library_id, title, folder_path, mod_time, duration_seconds)
values (?, ?, ?, ?, ?, ?)
returning *
;

-- name: GetAudioByID :one
select *
from audio
where id = ?
;

-- name: GetAudioPathsByLibrary :many
select id, folder_path, mod_time
from audio
where library_id = ?
;

-- name: UpdateAudioMeta :exec
update audio
set mod_time = ?,
    duration_seconds = ?,
    updated_at = datetime('now')
where id = ?
;

-- name: TouchAudio :exec
update audio
set updated_at = datetime('now')
where id = ?
;

-- name: UpdateAudio :exec
update audio
set title = ?,
    release_date = ?,
    language = ?,
    updated_at = datetime('now')
where id = ?
;

-- name: UpdateAudioCoverIndex :exec
-- cover_index is presentation state, not metadata - deliberately its own
-- query rather than folded into UpdateAudio, so a metadata edit can never
-- clobber it. updated_at is left untouched on purpose: picking a cover is
-- not a content change and must not resurface the work in "recently
-- updated".
update audio
set cover_index = ?
where id = ?
;

-- name: ListAudioIDsByLibrary :many
select id
from audio
where library_id = ?
;

-- name: DeleteAudio :exec
delete from audio
where id = ?
;

-- name: DeleteAudioFiles :exec
delete from audio_file
where audio_id = ?
;

-- name: UpdateAudioFileSortOrder :exec
update audio_file
set sort_order = ?
where audio_id = ? and relative_path = ?
;

-- name: DeleteAudioImages :exec
delete from audio_image
where audio_id = ?
;

-- name: CreateAudioFile :exec
insert into audio_file (audio_id, relative_path, duration_seconds, sort_order)
values (?, ?, ?, ?)
;

-- name: CreateAudioImage :exec
insert into audio_image (audio_id, relative_path, sort_order)
values (?, ?, ?)
;

-- name: GetAudioFiles :many
select *
from audio_file
where audio_id = ?
order by sort_order
;

-- name: GetAudioImages :many
select *
from audio_image
where audio_id = ?
order by sort_order
;

-- name: GetAudioArtistIDs :many
select artist.id
from audio
join audio_artist on audio.id = audio_artist.audio_id
join artist on audio_artist.artist_id = artist.id
where audio.id = ?
;

-- name: GetAudioArtistNames :many
select artist.name
from audio
join audio_artist on audio.id = audio_artist.audio_id
join artist on audio_artist.artist_id = artist.id
where audio.id = ?
order by artist.name
;

-- name: BulkAddAudioArtists :exec
insert or ignore into audio_artist (audio_id, artist_id)
select ?, value
from json_each(sqlc.arg('artists'))
;

-- name: BulkGetAudioArtists :many
select id, name
from artist
where name in (sqlc.slice('artists'))
;

-- name: BulkRemoveAudioArtists :exec
delete from audio_artist
where audio_id = ? and artist_id in (select value from json_each(sqlc.arg('artists')))
;

-- name: RepointAudioArtistLinks :exec
insert or ignore into audio_artist (audio_id, artist_id)
select audio_id, cast(sqlc.arg('target_id') as integer)
from audio_artist
where artist_id = cast(sqlc.arg('source_id') as integer)
;

-- name: GetAudioTagIDs :many
select tag.id
from audio
join audio_tag on audio.id = audio_tag.audio_id
join tag on audio_tag.tag_id = tag.id
where audio.id = ?
;

-- name: GetAudioTagNames :many
select tag.name
from audio
join audio_tag on audio.id = audio_tag.audio_id
join tag on audio_tag.tag_id = tag.id
where audio.id = ?
order by tag.name
;

-- name: BulkAddAudioTags :exec
insert or ignore into audio_tag (audio_id, tag_id)
select ?, value
from json_each(sqlc.arg('tags'))
;

-- name: BulkGetAudioTags :many
select id, name
from tag
where name in (sqlc.slice('tags'))
;

-- name: BulkRemoveAudioTags :exec
delete from audio_tag
where audio_id = ? and tag_id in (select value from json_each(sqlc.arg('tags')))
;

-- name: GetAudioCircleIDs :many
select circle.id
from audio
join audio_circle on audio.id = audio_circle.audio_id
join circle on audio_circle.circle_id = circle.id
where audio.id = ?
;

-- name: GetAudioCircleNames :many
select circle.name
from audio
join audio_circle on audio.id = audio_circle.audio_id
join circle on audio_circle.circle_id = circle.id
where audio.id = ?
order by circle.name
;

-- name: BulkAddAudioCircles :exec
insert or ignore into audio_circle (audio_id, circle_id)
select ?, value
from json_each(sqlc.arg('circles'))
;

-- name: BulkGetAudioCircles :many
select id, name
from circle
where name in (sqlc.slice('circles'))
;

-- name: BulkRemoveAudioCircles :exec
delete from audio_circle
where audio_id = ? and circle_id in (select value from json_each(sqlc.arg('circles')))
;

-- name: GetAudioVoiceActorIDs :many
select voice_actor.id
from audio
join audio_voice_actor on audio.id = audio_voice_actor.audio_id
join voice_actor on audio_voice_actor.voice_actor_id = voice_actor.id
where audio.id = ?
;

-- name: GetAudioVoiceActorNames :many
select voice_actor.name
from audio
join audio_voice_actor on audio.id = audio_voice_actor.audio_id
join voice_actor on audio_voice_actor.voice_actor_id = voice_actor.id
where audio.id = ?
order by voice_actor.name
;

-- name: BulkAddAudioVoiceActors :exec
insert or ignore into audio_voice_actor (audio_id, voice_actor_id)
select ?, value
from json_each(sqlc.arg('voice_actors'))
;

-- name: BulkGetAudioVoiceActors :many
select id, name
from voice_actor
where name in (sqlc.slice('voice_actors'))
;

-- name: BulkRemoveAudioVoiceActors :exec
delete from audio_voice_actor
where audio_id = ? and voice_actor_id in (select value from json_each(sqlc.arg('voice_actors')))
;

-- name: AddFavoriteAudio :exec
insert into audio_favorite (audio_id, user_id, favorited_at)
values (?, ?, datetime('now'))
on conflict (audio_id, user_id) do nothing
;

-- name: RemoveFavoriteAudio :exec
delete from audio_favorite
where audio_id = ? and user_id = sqlc.arg('uid')
;

-- name: AudioIsFavorited :one
select exists (
    select 1 from audio_favorite where audio_id = ? and user_id = sqlc.arg('uid')
) = 1
;

-- name: AddFavoriteAudioImage :exec
insert into audio_image_favorite (audio_id, user_id, image_index, favorited_at)
values (sqlc.arg('audio_id'), sqlc.arg('uid'), sqlc.arg('image_index'), datetime('now'))
on conflict (audio_id, user_id, image_index) do nothing
;

-- name: RemoveFavoriteAudioImage :exec
delete from audio_image_favorite
where audio_id = sqlc.arg('audio_id')
  and user_id = sqlc.arg('uid')
  and image_index = sqlc.arg('image_index')
;

-- name: GetFavoriteAudioImageIndices :many
select image_index
from audio_image_favorite
where audio_id = sqlc.arg('audio_id') and user_id = sqlc.arg('uid')
order by image_index asc
;

-- name: GetFavoritedAudioIDs :many
select audio_id
from audio_favorite
where user_id = sqlc.arg('uid') and audio_id in (sqlc.slice('ids'))
;

-- name: UpsertAudioRating :one
insert into audio_rating (audio_id, user_id, rating, rated_at)
values (?, ?, ?, datetime('now'))
on conflict (audio_id, user_id) do update set
    rating = excluded.rating,
    rated_at = excluded.rated_at
returning *
;

-- name: RemoveAudioRating :exec
delete from audio_rating
where audio_id = ? and user_id = sqlc.arg('uid')
;

-- name: GetAudioRating :one
select rating
from audio_rating
where audio_id = ? and user_id = sqlc.arg('uid')
;

-- name: GetAudioRatingsForIDs :many
select audio_id, rating
from audio_rating
where user_id = sqlc.arg('uid') and audio_id in (sqlc.slice('ids'))
;

-- name: MarkAudioStarted :exec
-- Records the latest playback start without changing completion. The insert
-- defaults completed to false; the conflict branch deliberately leaves it
-- alone so replaying a finished work keeps its finished state.
insert into audio_played (audio_id, user_id, played_at)
values (?, ?, datetime('now'))
on conflict (audio_id, user_id) do update set played_at = excluded.played_at
;

-- name: MarkAudioPlayed :exec
-- Explicit "mark as finished" and automatic end-of-queue tracking both use
-- this: the work is finished and its played_at is the completion time.
insert into audio_played (audio_id, user_id, played_at, completed)
values (?, ?, datetime('now'), 1)
on conflict (audio_id, user_id) do update set
    played_at = excluded.played_at,
    completed = 1
;

-- name: MarkAudioUnplayed :exec
delete from audio_played
where audio_id = ? and user_id = sqlc.arg('uid')
;

-- name: GetAudioPlayState :one
select completed
from audio_played
where audio_id = ? and user_id = sqlc.arg('uid')
;

-- name: GetAudioPlayStates :many
select audio_id, completed
from audio_played
where user_id = sqlc.arg('uid') and audio_id in (sqlc.slice('ids'))
;

-- name: GetAudiosByIDs :many
select id, library_id, title, release_date, language
from audio
where id in (sqlc.slice('ids'))
;

-- name: MarkAudioPlayedBulk :exec
insert into audio_played (audio_id, user_id, played_at, completed)
select audio.id, sqlc.arg('uid'), datetime('now'), 1
from audio
where audio.id in (sqlc.slice('ids'))
on conflict (audio_id, user_id) do update set
    played_at = excluded.played_at,
    completed = 1
;

-- name: MarkAudioUnplayedBulk :exec
delete from audio_played
where user_id = sqlc.arg('uid') and audio_id in (sqlc.slice('ids'))
;

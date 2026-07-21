-- name: GetReaderSettings :one
select *
from reader_settings
where user_id = ?
;

-- name: UpsertReaderSettings :one
insert into reader_settings (
  user_id,
  reading_direction,
  page_layout,
  fit_mode,
  background,
  view_mode
) values ( ?, ?, ?, ?, ?, ? )
on conflict (user_id) do update set
  reading_direction = excluded.reading_direction,
  page_layout = excluded.page_layout,
  fit_mode = excluded.fit_mode,
  background = excluded.background,
  view_mode = excluded.view_mode
returning *
;

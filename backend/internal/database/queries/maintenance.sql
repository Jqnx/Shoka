-- name: GetMaintenanceTasks :many
select * from maintenance_task order by task;

-- ListMaintenanceTasksDue returns every maintenance task whose fixed interval
-- has elapsed: it has never run or the interval has passed since the last
-- attempt. interval_minutes is set by migration 00039 and has no write path,
-- so the > 0 guard only keeps a hand-edited 0 from scheduling.
-- Polled by the maintenance manager's scheduler.
-- name: ListMaintenanceTasksDue :many
select * from maintenance_task
where interval_minutes > 0
  and (
    last_run_at is null
    or datetime(last_run_at, '+' || interval_minutes || ' minutes') <= datetime('now')
  )
order by task;

-- MarkMaintenanceTaskRun records a run attempt. Deliberately does not touch
-- updated_at, which tracks migration edits to the fixed schedule rather than
-- run activity.
-- name: MarkMaintenanceTaskRun :exec
update maintenance_task
set last_run_at = datetime('now')
where task = ?;

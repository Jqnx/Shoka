-- name: GetMaintenanceTasks :many
select * from maintenance_task order by task;

-- name: GetMaintenanceTask :one
select * from maintenance_task where task = ?;

-- name: UpdateMaintenanceTaskInterval :exec
update maintenance_task
set interval_minutes = ?, updated_at = datetime('now')
where task = ?;

-- ListMaintenanceTasksDue returns every scheduled maintenance task whose
-- interval has elapsed: scheduling is enabled (interval > 0) and either it
-- has never run or the interval has elapsed since the last run attempt.
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
-- updated_at, which tracks configuration edits rather than run activity.
-- name: MarkMaintenanceTaskRun :exec
update maintenance_task
set last_run_at = datetime('now')
where task = ?;

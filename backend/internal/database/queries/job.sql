-- name: EnqueueJob :exec
insert into job (type, payload) 
values (?, ?)
returning *;

-- name: EnqueueJobAfter :exec
insert into job (type, payload, run_after) 
values (?, ?, ?)
returning *;

-- name: ClaimJob :one
update job
set status = 'running',
    attempts = attempts + 1,
    updated_at = datetime('now')
where id = (
  select id
  from job
  where status = 'pending'
  and run_after <= datetime('now')
  order by created_at asc
  limit 1
)
returning id, type, payload, attempts, max_attempts;

-- name: MarkJobAsDone :exec
update job
set status = 'done', 
    updated_at = datetime('now')
where id = ?;

-- name: MarkJobAsFailed :exec
update job
set status = 'failed', 
    error = ?, 
    updated_at = datetime('now')
where id = ?;

-- name: RequeueJob :exec
update job
set status = 'pending',
    run_after = ?,
    updated_at = datetime('now')
where id = ?;

-- name: HasPendingJob :one
select
    exists (select 1 from job where type = ? and status in ('pending', 'running')) = 1
;

-- name: GetJobsByStatus :many
select
    id, type, status, attempts, max_attempts, error, created_at, updated_at, run_after
from job
where status = ?
order by created_at desc
;

-- name: GetJobCounts :many
select status, count(*) as count
from job
group by status
;

-- name: DeleteOldJobs :exec
delete from job
where status in ('failed', 'done') and updated_at < datetime('now', ?)  -- Takes a relative time string as parameter
;

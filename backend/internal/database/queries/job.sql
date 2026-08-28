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
  -- datetime() on both sides rather than a bare string comparison: Go
  -- writes time.Time as local time with a UTC offset ("...14:52:54+02:00")
  -- while datetime('now') is UTC, so comparing the raw text compares
  -- wall-clock digits from different zones and delays (or, in a negative
  -- offset, skips) every backoff.
  and datetime(run_after) <= datetime('now')
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
-- Scoped to type AND payload: two jobs of the same type but for different
-- targets (e.g. a scan for library A vs library B, or a thumbnail job for
-- archive X vs archive Y) must not suppress each other.
select
    exists (
        select 1 from job
        where type = sqlc.arg('type')
        and payload = sqlc.arg('payload')
        and status in ('pending', 'running')
    ) = 1
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

-- name: GetActiveJobs :many
-- Powers the admin job monitor. Selects payload (unlike GetJobsByStatus) so
-- the handler can derive a human-readable target, and is bounded so a
-- library-wide backfill of thousands of jobs can't blow up the response.
-- Running first, then oldest-queued first, matching claim order.
select
    id, type, payload, status, attempts, max_attempts, error, created_at, updated_at, run_after
from job
where status in ('pending', 'running')
order by
    case status when 'running' then 0 else 1 end,
    created_at asc
limit sqlc.arg('limit')
;

-- name: GetRecentFailedJobs :many
-- Most recently failed first - updated_at is when the failure was recorded.
select
    id, type, payload, status, attempts, max_attempts, error, created_at, updated_at, run_after
from job
where status = 'failed'
order by updated_at desc
limit sqlc.arg('limit')
;

-- name: ResetRunningJobs :execrows
-- Startup recovery: a job left 'running' by an unclean shutdown would
-- otherwise stay that way forever and show as permanently stuck in the
-- monitor. Back to 'pending' rather than 'failed' because ClaimJob already
-- incremented attempts, so a recovered job retries under the normal
-- max_attempts budget and fails on its own if it keeps dying.
update job
set status = 'pending',
    updated_at = datetime('now')
where status = 'running'
;

-- name: DeleteOldJobs :exec
delete from job
where status in ('failed', 'done') and updated_at < datetime('now', ?)  -- Takes a relative time string as parameter
;

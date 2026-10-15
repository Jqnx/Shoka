package jobs

import (
	"Shoka/internal/database/sqlc"
	"context"
	"encoding/json"
)

// Job is a queued unit of work as stored in the jobs table: its identity,
// type, raw JSON payload, and retry counters.
type Job struct {
	ID          int64
	Type        string
	RawPayload  string
	Attempts    int64
	MaxAttempts int64
}

// Decode unmarshals the job's raw JSON payload into target.
func (j *Job) Decode(target any) error {
	return json.Unmarshal([]byte(j.RawPayload), target)
}

// archiveGone reports whether the archive a job targets has been deleted. Jobs
// that write per-archive state (cache dirs, hashes) poll it: a delete evicts
// the archive's cache directory, so a job that kept working would recreate it
// as an orphan.
func archiveGone(ctx context.Context, q *sqlc.Queries, archiveID string) (bool, error) {
	exists, err := q.HasArchive(ctx, archiveID)
	if err != nil {
		return false, err
	}

	return !exists, nil
}

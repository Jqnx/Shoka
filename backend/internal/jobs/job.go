package jobs

import "encoding/json"

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

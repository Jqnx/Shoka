package jobs

import "encoding/json"

type Job struct {
	ID          int64
	Type        string
	RawPayload  string
	Attempts    int64
	MaxAttempts int64
}

func (j *Job) Decode(target any) error {
	return json.Unmarshal([]byte(j.RawPayload), target)
}

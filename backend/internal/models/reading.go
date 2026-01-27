package models

import "time"

type ReadingState int

const (
	StateNoUser ReadingState = iota
	StateUnread
	StateReading
	StateFinished
)

var StateName = map[ReadingState]string{
	StateNoUser:   "no user",
	StateUnread:   "unread",
	StateReading:  "reading",
	StateFinished: "finished",
}

type ReadingProgress struct {
	Status   ReadingState
	Progress int16
	LastRead *time.Time
}

func (s ReadingState) String() string {
	return StateName[s]
}

func (rp *ReadingProgress) SetState(state string) {
	for r, s := range StateName {
		if state == s {
			rp.Status = r
			return
		}
	}
}

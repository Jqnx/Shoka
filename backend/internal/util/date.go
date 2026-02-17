package util

import "time"

func DateToTime(day, month, year int) *time.Time {
	if day == 0 || month == 0 || year == 0 {
		return nil
	}
	date := time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.UTC)
	return &date
}

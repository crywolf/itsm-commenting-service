package testutils

import "time"

// FixedClock is used in testing. It implements Clock interface used in mock storages to create timestamp.
type FixedClock struct{}

// Now returns fixed time
func (FixedClock) Now() time.Time {
	tz, err := time.LoadLocation("Europe/Prague")
	if err != nil {
		panic(err)
	}
	return time.Date(2021, 4, 1, 12, 34, 56, 78, tz)
}

// NowFormatted returns fixed time string in RFC3339 format
func (c FixedClock) NowFormatted() string {
	return c.Now().Format(time.RFC3339)
}

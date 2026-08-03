package main

import "time"

// dateRange lists every day from newest down to oldest, inclusive. The walk goes
// backwards so a short run covers the most recent days first, which lets a small
// --days value act as a rehearsal for a larger one.
func dateRange(newest, oldest time.Time) []time.Time {
	var out []time.Time
	for d := newest; !d.Before(oldest); d = d.AddDate(0, 0, -1) {
		out = append(out, d)
	}
	return out
}

// daysAgoRange lists n days ending at (and including) end.
func daysAgoRange(end time.Time, n int) []time.Time {
	if n < 1 {
		return nil
	}
	return dateRange(end, end.AddDate(0, 0, -(n-1)))
}

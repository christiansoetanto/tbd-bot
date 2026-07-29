package main

import (
	"testing"
	"time"
)

func day(y int, m time.Month, d int) time.Time {
	return time.Date(y, m, d, 0, 0, 0, 0, time.UTC)
}

func TestDateRange_WalksBackwardsFromMostRecent(t *testing.T) {
	// Backfill starts at the most recent missing day and works backwards, so a
	// short run verifies the newest days before committing to a long one.
	got := dateRange(day(2026, time.July, 28), day(2026, time.July, 25))

	want := []time.Time{
		day(2026, time.July, 28),
		day(2026, time.July, 27),
		day(2026, time.July, 26),
		day(2026, time.July, 25),
	}
	if len(got) != len(want) {
		t.Fatalf("got %d dates, want %d", len(got), len(want))
	}
	for i := range want {
		if !got[i].Equal(want[i]) {
			t.Errorf("index %d: got %s, want %s", i, got[i].Format("2006-01-02"), want[i].Format("2006-01-02"))
		}
	}
}

func TestDateRange_SingleDay(t *testing.T) {
	got := dateRange(day(2026, time.July, 28), day(2026, time.July, 28))
	if len(got) != 1 || !got[0].Equal(day(2026, time.July, 28)) {
		t.Fatalf("expected exactly the one day, got %v", got)
	}
}

func TestDateRange_EmptyWhenOldestAfterNewest(t *testing.T) {
	got := dateRange(day(2026, time.July, 25), day(2026, time.July, 28))
	if len(got) != 0 {
		t.Fatalf("expected no dates when the range is inverted, got %d", len(got))
	}
}

// The staged rollout the operator asked for -- 1 day, then 7, then 30 -- is
// expressed as --days, so that must line up with a plain backwards walk.
func TestDaysAgoRange(t *testing.T) {
	end := day(2026, time.July, 28)

	one := daysAgoRange(end, 1)
	if len(one) != 1 || !one[0].Equal(end) {
		t.Errorf("--days 1 should yield only the end date, got %v", one)
	}

	seven := daysAgoRange(end, 7)
	if len(seven) != 7 {
		t.Fatalf("--days 7 should yield 7 dates, got %d", len(seven))
	}
	if !seven[0].Equal(end) {
		t.Errorf("first date should be the end date, got %s", seven[0].Format("2006-01-02"))
	}
	if want := day(2026, time.July, 22); !seven[6].Equal(want) {
		t.Errorf("last date = %s, want %s", seven[6].Format("2006-01-02"), want.Format("2006-01-02"))
	}
}

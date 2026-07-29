package util

import (
	"strings"
	"testing"
	"time"
)

// Mirrors the structure the parser relies on: a .rubrica marker whose parent
// holds the reading, with a RESPONSORY marker that must be cut off.
const fixtureHTML = `
<html><body>
<div class="lettura">
  <span class="rubrica">SECOND READING</span>
  From a treatise by Saint Example
  <p>The body of the reading goes here.</p>
  <span class="rubrica">RESPONSORY</span>
  <p>Responsory text that must not appear.</p>
</div>
</body></html>`

func TestParseOfficeOfReadings_UsesSuppliedDateNotToday(t *testing.T) {
	date := time.Date(2025, time.September, 21, 0, 0, 0, 0, time.UTC)

	title, text, err := ParseOfficeOfReadings(strings.NewReader(fixtureHTML), date)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	want := "Office of Readings for Sunday, 21 September 2025"
	if title != want {
		t.Errorf("title = %q, want %q", title, want)
	}
	if !strings.Contains(text, "The body of the reading goes here.") {
		t.Errorf("reading body missing from text: %q", text)
	}
	if strings.Contains(text, "Responsory text that must not appear.") {
		t.Errorf("responsory was not stripped: %q", text)
	}
	if strings.Contains(text, "SECOND READING") {
		t.Errorf("SECOND READING marker was not stripped: %q", text)
	}
}

func TestParseOfficeOfReadings_ErrorsWhenNoSecondReading(t *testing.T) {
	const noReading = `<html><body><div><span class="rubrica">INVITATORY</span></div></body></html>`

	_, _, err := ParseOfficeOfReadings(strings.NewReader(noReading), time.Now())
	if err == nil {
		t.Fatal("expected an error when the page has no SECOND READING section")
	}
}

// The backfill writes one file per date, and those filenames must match what
// the daily cron produces or the same day gets two differently-named files.
func TestOfficeOfReadingsFilename(t *testing.T) {
	date := time.Date(2025, time.September, 21, 0, 0, 0, 0, time.UTC)

	got := OfficeOfReadingsFilename(date)
	want := "2025-09-21-Office of Readings for Sunday, 21 September 2025.md"
	if got != want {
		t.Errorf("filename = %q, want %q", got, want)
	}
}

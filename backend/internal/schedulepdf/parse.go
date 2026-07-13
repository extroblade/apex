// Package schedulepdf ingests iRacing season-schedule PDFs: it finds candidate
// PDFs (probed URLs + a local drop directory), extracts their text, parses the
// per-series week/track tables, and upserts the result into season_schedule.
//
// The text-layout parser is unit-tested against the known PDF table format;
// extraction from a live iRacing PDF should be validated once a real file is
// available (the download is not publicly reachable without an account).
package schedulepdf

import (
	"regexp"
	"strconv"
	"strings"
)

// SeriesSchedule is one parsed series with its week -> track table.
type SeriesSchedule struct {
	SeriesName string
	Weeks      map[int]string // week number -> track name (as printed)
}

// weekLine matches schedule rows like:
//
//	"Week 1 (2026-06-16) Circuit de Spa-Francorchamps - Grand Prix"
//	"Week 10 Okayama International Circuit"
var weekLine = regexp.MustCompile(`^Week\s+(\d{1,2})\s*(?:\([^)]*\))?\s*[-–:]?\s*(.+?)\s*$`)

// seriesHeader matches lines that start a new series section. iRacing prints
// the series name as a standalone line followed by its schedule table; we
// treat any non-week, non-noise line as a potential header and confirm it when
// week rows follow.
var noiseLine = regexp.MustCompile(`(?i)^(page \d+|iracing.*schedule|20\d\d season.*|week\s*$|track\s*$|date\s*$|car[s]?\s*$|license\s*$|\s*)$`)

// ParseText parses extracted PDF text lines into per-series schedules. A series
// is recognized as a header line followed (possibly after noise) by "Week N"
// rows; the series ends when a new header collects its own week rows.
func ParseText(lines []string) []SeriesSchedule {
	var (
		out     []SeriesSchedule
		current *SeriesSchedule
		pending string // last seen candidate header
	)

	flush := func() {
		if current != nil && len(current.Weeks) > 0 {
			out = append(out, *current)
		}
		current = nil
	}

	for _, raw := range lines {
		line := strings.TrimSpace(raw)
		if noiseLine.MatchString(line) {
			continue
		}

		if m := weekLine.FindStringSubmatch(line); m != nil {
			week, err := strconv.Atoi(m[1])
			track := strings.TrimSpace(m[2])
			if err != nil || week < 1 || week > 13 || track == "" {
				continue
			}
			if current == nil {
				if pending == "" {
					continue // week row with no series context; skip
				}
				current = &SeriesSchedule{SeriesName: pending, Weeks: map[int]string{}}
				pending = ""
			}
			current.Weeks[week] = track
			continue
		}

		// A non-week content line: candidate for the next series header.
		if current != nil {
			flush()
		}
		pending = line
	}
	flush()
	return out
}

// Normalize lowercases and strips punctuation/spacing so names from the PDF
// can be matched against catalog names.
func Normalize(s string) string {
	s = strings.ToLower(s)
	var b strings.Builder
	for _, r := range s {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			b.WriteRune(r)
		}
	}
	return b.String()
}

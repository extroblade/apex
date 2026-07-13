package schedulepdf

import "testing"

func TestParseText(t *testing.T) {
	lines := []string{
		"iRacing 2026 Season 3 Schedule",
		"Page 1",
		"",
		"Global Mazda MX-5 Fanatec Cup",
		"License",
		"Week 1 (2026-06-16) Okayama International Circuit",
		"Week 2 (2026-06-23) Circuit de Spa-Francorchamps - Grand Prix",
		"Week 3 Summit Point Raceway",
		"",
		"GT Sprint",
		"Week 1 (2026-06-16) Circuit de Barcelona-Catalunya",
		"Week 2: Mount Panorama Circuit",
		"Week 13 (2026-09-08) Suzuka International Racing Course",
	}

	got := ParseText(lines)
	if len(got) != 2 {
		t.Fatalf("want 2 series, got %d: %+v", len(got), got)
	}

	mx5 := got[0]
	if mx5.SeriesName != "Global Mazda MX-5 Fanatec Cup" || len(mx5.Weeks) != 3 {
		t.Fatalf("mx5 parse wrong: %+v", mx5)
	}
	if mx5.Weeks[2] != "Circuit de Spa-Francorchamps - Grand Prix" {
		t.Errorf("mx5 week 2: %q", mx5.Weeks[2])
	}

	gts := got[1]
	if gts.SeriesName != "GT Sprint" {
		t.Fatalf("second series: %+v", gts)
	}
	if gts.Weeks[2] != "Mount Panorama Circuit" || gts.Weeks[13] != "Suzuka International Racing Course" {
		t.Errorf("gt sprint weeks wrong: %+v", gts.Weeks)
	}
}

func TestParseTextIgnoresOrphanWeeks(t *testing.T) {
	got := ParseText([]string{"Week 1 Some Track", "Week 2 Another"})
	if len(got) != 0 {
		t.Fatalf("orphan week rows must not create a series: %+v", got)
	}
}

func TestNormalize(t *testing.T) {
	a := Normalize("Circuit de Spa-Francorchamps - Grand Prix")
	b := Normalize("circuit de spa francorchamps grand prix")
	if a != b {
		t.Errorf("normalize mismatch: %q vs %q", a, b)
	}
}

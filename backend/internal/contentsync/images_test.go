package contentsync

import (
	"strings"
	"testing"
)

func TestCapDescription(t *testing.T) {
	short := "A concise blurb."
	if got := capDescription(short); got != short {
		t.Errorf("short text changed: %q", got)
	}

	long := strings.Repeat("word ", 500) // 2500 runes, well over the cap
	got := capDescription(long)
	if n := len([]rune(got)); n > descriptionMaxLen+1 { // +1 for the ellipsis
		t.Errorf("capped length = %d runes, want <= %d", n, descriptionMaxLen+1)
	}
	if !strings.HasSuffix(got, "…") {
		t.Errorf("expected ellipsis suffix, got %q", got[len(got)-8:])
	}
	if strings.HasSuffix(strings.TrimSuffix(got, "…"), " ") {
		t.Errorf("should trim trailing space before ellipsis: %q", got)
	}
}

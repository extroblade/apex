package setups

import (
	"strings"
	"testing"
)

func TestVariationDeterministicAndBounded(t *testing.T) {
	for car := 1; car < 50; car++ {
		for track := 0; track < 50; track++ {
			v := variation(car, track, 3)
			if v < -3 || v > 3 {
				t.Fatalf("variation(%d,%d,3) = %d out of [-3,3]", car, track, v)
			}
			if v != variation(car, track, 3) {
				t.Fatalf("variation not deterministic for (%d,%d)", car, track)
			}
		}
	}
	if variation(1, 2, 0) != 0 {
		t.Error("zero spread must yield zero offset")
	}
}

func TestTweakFor(t *testing.T) {
	tests := []struct {
		cat, name, config string
		wantLabel         string
	}{
		{"road", "Nürburgring Combined", "Gesamtstrecke 24h", "long lap"},
		{"road", "Autodromo Nazionale Monza", "Grand Prix", "low-drag"},
		{"oval", "Daytona International Speedway", "Oval", "oval"},
		{"dirt_oval", "Eldora Speedway", "", "loose surface"},
		{"road", "Okayama International Circuit", "Full", "neutral"},
	}
	for _, tt := range tests {
		got := tweakFor(tt.cat, tt.name, tt.config)
		if !strings.Contains(got.label, tt.wantLabel) {
			t.Errorf("tweakFor(%s, %s): label %q, want containing %q", tt.cat, tt.name, got.label, tt.wantLabel)
		}
	}
}

func TestBaselinesCoverAllCategories(t *testing.T) {
	for _, cat := range []string{"formula_car", "sports_car", "oval", "dirt_oval", "dirt_road"} {
		b, ok := baselines[cat]
		if !ok {
			t.Fatalf("no baseline for category %s", cat)
		}
		if b.tirePsi <= 0 || b.springN <= 0 || b.notes == "" {
			t.Errorf("baseline %s has zero fields: %+v", cat, b)
		}
	}
}

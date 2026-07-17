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

// TestPackMatrix checks the pack is the full 2×4 grid and every axis it names
// has both a delta and display metadata.
func TestPackMatrix(t *testing.T) {
	if len(packOrder) != 8 {
		t.Fatalf("packOrder has %d entries, want 8 (2 skills × 4 sessions)", len(packOrder))
	}
	seen := map[string]bool{}
	for _, v := range packOrder {
		if _, ok := skillDeltas[v.skill]; !ok {
			t.Errorf("no skill delta for %q", v.skill)
		}
		if _, ok := skillMeta[v.skill]; !ok {
			t.Errorf("no skill meta for %q", v.skill)
		}
		if _, ok := sessionDeltas[v.session]; !ok {
			t.Errorf("no session delta for %q", v.session)
		}
		if _, ok := sessionMeta[v.session]; !ok {
			t.Errorf("no session meta for %q", v.session)
		}
		key := v.skill + "/" + v.session
		if seen[key] {
			t.Errorf("duplicate variant %q", key)
		}
		seen[key] = true
	}
}

// TestComputeVariantsSemantics locks in the intended direction of each axis on a
// GT baseline + neutral track, so the deltas can't be silently inverted.
func TestComputeVariantsSemantics(t *testing.T) {
	b := baselineFor("sports_car")
	tw := tweakFor("road", "Okayama", "Full")
	got := func(skill, session string) computed {
		d := addDelta(skillDeltas[skill], sessionDeltas[session])
		return computeSetup(b, tw, d, 7, 3, 1) // fixed ids/salt → deterministic
	}

	// Pro is stiffer and more front-biased than Safe.
	if got("pro", "race").spring <= got("safe", "race").spring {
		t.Errorf("pro should be stiffer than safe")
	}
	if got("pro", "race").bias <= got("safe", "race").bias {
		t.Errorf("pro should carry more front brake bias than safe")
	}
	// Rain runs much more wing than qualifying (same skill).
	if got("safe", "rain").wingR <= got("safe", "qual").wingR {
		t.Errorf("rain should run more rear wing than qual")
	}
	// Rain sits higher and softer than a dry race setup.
	if got("safe", "rain").ride <= got("safe", "race").ride {
		t.Errorf("rain should sit higher than race")
	}
	if got("safe", "rain").spring >= got("safe", "race").spring {
		t.Errorf("rain should be softer than race")
	}
}

// TestComputeOvalGuards ensures skill/session deltas don't sprout wing or diff
// settings on a discipline that has none (an oval).
func TestComputeOvalGuards(t *testing.T) {
	b := baselineFor("oval")
	tw := tweakFor("oval", "Daytona", "Oval")
	for _, v := range packOrder {
		d := addDelta(skillDeltas[v.skill], sessionDeltas[v.session])
		c := computeSetup(b, tw, d, 12, 8, 2)
		if c.wingR != 0 || c.wingF != 0 {
			t.Errorf("oval variant %s/%s grew a wing: %+v", v.skill, v.session, c)
		}
		if c.diff != 0 {
			t.Errorf("oval variant %s/%s grew a diff preload: %+v", v.skill, v.session, c)
		}
	}
}

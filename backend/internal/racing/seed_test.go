package racing

import (
	"encoding/json"
	"testing"
)

// The embedded seed is generated (scripts/gen-catalog-seed.py) — verify its
// integrity so a bad regeneration can't ship: real ids only, weeks within the
// season, and every schedule row pointing at a track that exists.
func TestCatalogSeedIntegrity(t *testing.T) {
	var data seedData
	if err := json.Unmarshal(catalogSeed, &data); err != nil {
		t.Fatalf("seed does not parse: %v", err)
	}
	if len(data.Cars) < 100 || len(data.Tracks) < 300 || len(data.Series) < 100 {
		t.Fatalf("seed suspiciously small: %d cars, %d tracks, %d series",
			len(data.Cars), len(data.Tracks), len(data.Series))
	}

	trackIDs := make(map[int]bool, len(data.Tracks))
	for _, tr := range data.Tracks {
		trackIDs[tr.TrackID] = true
	}
	carIDs := make(map[int]bool, len(data.Cars))
	freeCars := 0
	for _, c := range data.Cars {
		carIDs[c.CarID] = true
		if c.Free {
			freeCars++
		}
	}
	if freeCars == 0 {
		t.Error("no free cars in seed — free flags missing")
	}

	scheduled := 0
	for _, s := range data.Series {
		for _, w := range s.Weeks {
			if w.Week < 1 || w.Week > SeasonWeeks {
				t.Errorf("series %d week %d out of 1..%d", s.SeriesID, w.Week, SeasonWeeks)
			}
			if !trackIDs[w.TrackID] {
				t.Errorf("series %d week %d references unknown track %d", s.SeriesID, w.Week, w.TrackID)
			}
			if len(w.Date) != 10 {
				t.Errorf("series %d week %d has bad date %q", s.SeriesID, w.Week, w.Date)
			}
			scheduled++
		}
		for _, carID := range s.Cars {
			if !carIDs[carID] {
				t.Errorf("series %d references unknown car %d", s.SeriesID, carID)
			}
		}
	}
	if scheduled < 500 {
		t.Errorf("only %d scheduled races — expected a full season", scheduled)
	}

	for trackID, reqs := range data.TrackRequirements {
		if len(reqs) == 0 {
			t.Errorf("track %s has empty requirements", trackID)
		}
		for _, r := range reqs {
			if !trackIDs[r] {
				t.Errorf("requirement for %s references unknown track %d", trackID, r)
			}
		}
	}
}

func TestNormalizeName(t *testing.T) {
	cases := map[string]string{
		"Nürburgring Grand-Prix-Strecke": "nurburgringgrandprixstrecke",
		"nurburgring grand prix strecke": "nurburgringgrandprixstrecke",
		"Global Mazda MX-5 Cup":          "globalmazdamx5cup",
	}
	for in, want := range cases {
		if got := normalizeName(in); got != want {
			t.Errorf("normalizeName(%q) = %q, want %q", in, got, want)
		}
	}
}

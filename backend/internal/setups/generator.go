package setups

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
)

// Generated is a baseline setup produced for a car+track combo. It is NOT
// telemetry-derived — it's a sane, deterministic starting point built from the
// car's discipline and the track's character, meant to be saved and tuned.
type Generated struct {
	Name  string `json:"name"`
	Notes string `json:"notes"`
	Data  string `json:"data"`
}

// baseline carries the tunable ranges for one discipline.
type baseline struct {
	tirePsi        float64 // cold pressure baseline, psi
	wingFront      int     // percent (formula/sports) — ovals use tape instead
	wingRear       int
	rideHeightMM   int
	springN        int // spring rate, N/mm
	arbFront       int // anti-roll bar setting 1..10
	arbRear        int
	brakeBias      float64 // % front
	diffPreloadNm  int
	notes          string
}

var baselines = map[string]baseline{
	"formula_car": {tirePsi: 21.0, wingFront: 28, wingRear: 32, rideHeightMM: 22, springN: 180, arbFront: 7, arbRear: 5, brakeBias: 56.5, diffPreloadNm: 40,
		notes: "High-downforce baseline: stiff, low, front bias forward for stability under braking."},
	"sports_car": {tirePsi: 23.5, wingFront: 3, wingRear: 6, rideHeightMM: 55, springN: 130, arbFront: 6, arbRear: 4, brakeBias: 54.0, diffPreloadNm: 60,
		notes: "GT baseline: balanced platform, softer than formula, forgiving on curbs."},
	"oval": {tirePsi: 28.0, wingFront: 0, wingRear: 0, rideHeightMM: 45, springN: 200, arbFront: 5, arbRear: 3, brakeBias: 58.0, diffPreloadNm: 0,
		notes: "Oval baseline: cross-weight and stagger carry the car — adjust wedge before springs."},
	"dirt_oval": {tirePsi: 12.0, wingFront: 0, wingRear: 0, rideHeightMM: 110, springN: 90, arbFront: 2, arbRear: 1, brakeBias: 50.0, diffPreloadNm: 0,
		notes: "Dirt oval baseline: soft and high; drive it loose, tighten as the track rubbers in."},
	"dirt_road": {tirePsi: 16.0, wingFront: 2, wingRear: 4, rideHeightMM: 95, springN: 100, arbFront: 3, arbRear: 2, brakeBias: 52.0, diffPreloadNm: 80,
		notes: "Rallycross baseline: compliance first — soft springs, generous ride height."},
}

// track-character adjustments, applied on top of the car baseline.
type trackTweak struct {
	psi, wing, ride, spring float64
	label                   string
}

func tweakFor(trackCategory, trackName, configName string) trackTweak {
	name := strings.ToLower(trackName + " " + configName)
	switch {
	case strings.Contains(name, "nordschleife") || strings.Contains(name, "combined"):
		return trackTweak{psi: +0.5, wing: -2, ride: +8, spring: -15, label: "long lap, bumpy — raised, softened, trimmed for the straights"}
	case strings.Contains(name, "monza") || strings.Contains(name, "le mans") || strings.Contains(name, "daytona") && trackCategory == "road":
		return trackTweak{psi: 0, wing: -6, ride: -2, spring: 0, label: "low-drag track — wings trimmed for top speed"}
	case strings.Contains(name, "monaco") || strings.Contains(name, "street") || strings.Contains(name, "long beach"):
		return trackTweak{psi: -0.5, wing: +4, ride: +4, spring: -10, label: "street circuit — max grip and compliance over curbs"}
	case trackCategory == "oval":
		return trackTweak{psi: +1.0, wing: 0, ride: -3, spring: +20, label: "oval — stiffer right side, pressures up for long runs"}
	case trackCategory == "dirt_oval" || trackCategory == "dirt_road":
		return trackTweak{psi: -1.0, wing: 0, ride: +10, spring: -10, label: "loose surface — softer and higher"}
	default:
		return trackTweak{psi: 0, wing: 0, ride: 0, spring: 0, label: "neutral road baseline"}
	}
}

// variation derives a small deterministic offset from the combo so different
// car+track pairs don't produce byte-identical files.
func variation(carID, trackID, spread int) int {
	if spread == 0 {
		return 0
	}
	h := uint32(carID*2654435761) ^ uint32(trackID*40503)
	return int(h%uint32(2*spread+1)) - spread
}

// Generate builds a baseline setup for the car+track combo, looking the pair
// up in the catalog for names and disciplines.
func (s *Service) Generate(ctx context.Context, carID, trackID int) (Generated, error) {
	var carName, carCat string
	err := s.db.QueryRowContext(ctx,
		`SELECT car_name, category FROM cars WHERE car_id = ?`, carID).Scan(&carName, &carCat)
	if errors.Is(err, sql.ErrNoRows) {
		return Generated{}, ErrInvalid
	}
	if err != nil {
		return Generated{}, err
	}

	trackName, configName, trackCat := "Any Track", "", "road"
	if trackID > 0 {
		err = s.db.QueryRowContext(ctx,
			`SELECT track_name, COALESCE(config_name, ''), category FROM tracks WHERE track_id = ?`,
			trackID).Scan(&trackName, &configName, &trackCat)
		if errors.Is(err, sql.ErrNoRows) {
			return Generated{}, ErrInvalid
		}
		if err != nil {
			return Generated{}, err
		}
	}

	b, ok := baselines[carCat]
	if !ok {
		b = baselines["sports_car"]
	}
	tw := tweakFor(trackCat, trackName, configName)

	psi := b.tirePsi + tw.psi + float64(variation(carID, trackID, 1))*0.25
	wingF := clamp(b.wingFront+int(tw.wing)/2, 0, 50)
	wingR := clamp(b.wingRear+int(tw.wing), 0, 50)
	ride := b.rideHeightMM + int(tw.ride) + variation(carID, trackID+1, 2)
	spring := b.springN + int(tw.spring) + variation(carID, trackID+2, 5)
	bias := b.brakeBias + float64(variation(carID, trackID+3, 2))*0.25

	trackLabel := trackName
	if configName != "" {
		trackLabel += " (" + configName + ")"
	}

	var sb strings.Builder
	fmt.Fprintf(&sb, "# %s @ %s — generated baseline\n\n", carName, trackLabel)
	fmt.Fprintf(&sb, "[TIRES]\n")
	fmt.Fprintf(&sb, "cold_pressure_lf = %.1f psi\ncold_pressure_rf = %.1f psi\n", psi, psi+ovalOffset(trackCat, 1.5))
	fmt.Fprintf(&sb, "cold_pressure_lr = %.1f psi\ncold_pressure_rr = %.1f psi\n\n", psi, psi+ovalOffset(trackCat, 1.5))
	fmt.Fprintf(&sb, "[SUSPENSION]\n")
	fmt.Fprintf(&sb, "ride_height = %d mm\nspring_rate = %d N/mm\narb_front = %d\narb_rear = %d\n\n",
		ride, spring, b.arbFront, b.arbRear)
	if wingR > 0 {
		fmt.Fprintf(&sb, "[AERO]\nwing_front = %d%%\nwing_rear = %d%%\n\n", wingF, wingR)
	}
	fmt.Fprintf(&sb, "[BRAKES]\nbrake_bias = %.2f%% front\n\n", bias)
	if b.diffPreloadNm > 0 {
		fmt.Fprintf(&sb, "[DIFFERENTIAL]\npreload = %d Nm\n\n", b.diffPreloadNm)
	}
	if trackCat == "oval" || trackCat == "dirt_oval" {
		fmt.Fprintf(&sb, "[OVAL]\ncross_weight = %.1f%%\nrear_stagger = %.2f in\n\n",
			50.0+float64(variation(carID, trackID+4, 4))*0.25, 0.25+float64(variation(carID, trackID+5, 2))*0.125)
	}
	fmt.Fprintf(&sb, "# %s\n# Track: %s\n", b.notes, tw.label)

	return Generated{
		Name:  fmt.Sprintf("%s @ %s — baseline", shorten(carName, 40), shorten(trackName, 40)),
		Notes: fmt.Sprintf("Generated baseline (%s; %s). Tune from here — it is a starting point, not a hotlap file.", carCat, tw.label),
		Data:  sb.String(),
	}, nil
}

func ovalOffset(cat string, v float64) float64 {
	if cat == "oval" || cat == "dirt_oval" {
		return v // right side runs higher pressure on ovals
	}
	return 0
}

func clamp(v, lo, hi int) int {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}

func shorten(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n-1] + "…"
}

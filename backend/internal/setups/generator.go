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

// GeneratedVariant is one setup in a generated pack: a Generated plus the axes
// it was built for. skill is "safe"|"pro"; session is
// "endurance"|"race"|"qual"|"rain".
type GeneratedVariant struct {
	Generated
	Skill   string `json:"skill"`
	Session string `json:"session"`
	Label   string `json:"label"` // human label, e.g. "Pro · Qualifying"
}

// baseline carries the tunable ranges for one discipline.
type baseline struct {
	tirePsi       float64 // cold pressure baseline, psi
	wingFront     int     // percent (formula/sports) — ovals use tape instead
	wingRear      int
	rideHeightMM  int
	springN       int // spring rate, N/mm
	arbFront      int // anti-roll bar setting 1..10
	arbRear       int
	brakeBias     float64 // % front
	diffPreloadNm int
	notes         string
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

func baselineFor(cat string) baseline {
	if b, ok := baselines[cat]; ok {
		return b
	}
	return baselines["sports_car"]
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

// delta is a set of offsets applied on top of the car baseline + track tweak.
// A skill delta and a session delta sum together to make one pack variant.
type delta struct {
	psi, wing, ride, spring, bias, diff float64
}

func addDelta(a, b delta) delta {
	return delta{a.psi + b.psi, a.wing + b.wing, a.ride + b.ride, a.spring + b.spring, a.bias + b.bias, a.diff + b.diff}
}

// skillDeltas trade stability for sharpness. "safe" is understeer-biased and
// forgiving; "pro" is stiffer, lower, and more front-biased.
var skillDeltas = map[string]delta{
	"safe": {psi: +0.5, wing: +3, ride: +3, spring: -12, bias: -1.0, diff: -10},
	"pro":  {psi: -0.3, wing: -2, ride: -3, spring: +15, bias: +1.0, diff: +15},
}

// sessionDeltas shape the setup for a session type. "race" is the neutral
// baseline (all zero); the others trade around it.
var sessionDeltas = map[string]delta{
	"endurance": {psi: +0.5, wing: +2, ride: +2, spring: -10, bias: -0.5, diff: -10},
	"race":      {},
	"qual":      {psi: -0.5, wing: -3, ride: -3, spring: +10, bias: +0.5, diff: +5},
	"rain":      {psi: -1.5, wing: +8, ride: +8, spring: -25, bias: -2.0, diff: -20},
}

type axisMeta struct{ name, note string }

var skillMeta = map[string]axisMeta{
	"safe": {"Safe", "forgiving and stable — understeer-biased"},
	"pro":  {"Pro", "sharp and responsive — rewards precise inputs"},
}

var sessionMeta = map[string]axisMeta{
	"endurance": {"Endurance", "tyre and fuel saving; softer and steadier for long stints"},
	"race":      {"Race", "balanced race-trim baseline"},
	"qual":      {"Qualifying", "one-lap pace; trimmed and stiffer for a single hot lap"},
	"rain":      {"Rain", "wet weather; more wing, higher, softer, gentler on throttle and brakes"},
}

// packOrder is the 2×4 matrix of variants a pack produces, most-used first.
var packOrder = []struct{ skill, session string }{
	{"safe", "race"}, {"pro", "race"},
	{"safe", "qual"}, {"pro", "qual"},
	{"safe", "endurance"}, {"pro", "endurance"},
	{"safe", "rain"}, {"pro", "rain"},
}

// variation derives a small deterministic offset from the combo so different
// car+track pairs (and pack salts) don't produce byte-identical files.
func variation(carID, trackID, spread int) int {
	if spread == 0 {
		return 0
	}
	h := uint32(carID*2654435761) ^ uint32(trackID*40503)
	return int(h%uint32(2*spread+1)) - spread
}

// computed holds the resolved numeric setup values ready to render.
type computed struct {
	psi          float64
	wingF, wingR int
	ride, spring int
	bias         float64
	arbF, arbR   int
	diff         int
}

// computeSetup resolves baseline + track tweak + delta into final values. salt
// varies the deterministic jitter so pack variants differ from each other. Wing
// and diff deltas apply only where the discipline actually has them (an oval has
// no wing, so a "+wing" delta must not sprout an AERO section).
func computeSetup(b baseline, tw trackTweak, d delta, carID, trackID, salt int) computed {
	c := computed{arbF: b.arbFront, arbR: b.arbRear}
	c.psi = b.tirePsi + tw.psi + d.psi + float64(variation(carID, trackID+salt, 1))*0.25
	if b.wingRear > 0 {
		c.wingF = clamp(b.wingFront+int((tw.wing+d.wing)/2), 0, 60)
		c.wingR = clamp(b.wingRear+int(tw.wing+d.wing), 0, 60)
	}
	c.ride = clamp(b.rideHeightMM+int(tw.ride+d.ride)+variation(carID, trackID+salt+1, 2), 5, 220)
	c.spring = clamp(b.springN+int(tw.spring+d.spring)+variation(carID, trackID+salt+2, 5), 40, 400)
	c.bias = b.brakeBias + d.bias + float64(variation(carID, trackID+salt+3, 2))*0.25
	if b.diffPreloadNm > 0 {
		c.diff = clamp(b.diffPreloadNm+int(d.diff), 0, 300)
	}
	return c
}

// render turns resolved values into the plain-text setup file the showroom stores.
func render(carName, trackLabel, trackCat string, c computed, carID, trackID int, footer string) string {
	var sb strings.Builder
	fmt.Fprintf(&sb, "# %s @ %s\n\n", carName, trackLabel)
	fmt.Fprintf(&sb, "[TIRES]\n")
	fmt.Fprintf(&sb, "cold_pressure_lf = %.1f psi\ncold_pressure_rf = %.1f psi\n", c.psi, c.psi+ovalOffset(trackCat, 1.5))
	fmt.Fprintf(&sb, "cold_pressure_lr = %.1f psi\ncold_pressure_rr = %.1f psi\n\n", c.psi, c.psi+ovalOffset(trackCat, 1.5))
	fmt.Fprintf(&sb, "[SUSPENSION]\n")
	fmt.Fprintf(&sb, "ride_height = %d mm\nspring_rate = %d N/mm\narb_front = %d\narb_rear = %d\n\n",
		c.ride, c.spring, c.arbF, c.arbR)
	if c.wingR > 0 {
		fmt.Fprintf(&sb, "[AERO]\nwing_front = %d%%\nwing_rear = %d%%\n\n", c.wingF, c.wingR)
	}
	fmt.Fprintf(&sb, "[BRAKES]\nbrake_bias = %.2f%% front\n\n", c.bias)
	if c.diff > 0 {
		fmt.Fprintf(&sb, "[DIFFERENTIAL]\npreload = %d Nm\n\n", c.diff)
	}
	if trackCat == "oval" || trackCat == "dirt_oval" {
		fmt.Fprintf(&sb, "[OVAL]\ncross_weight = %.1f%%\nrear_stagger = %.2f in\n\n",
			50.0+float64(variation(carID, trackID+4, 4))*0.25, 0.25+float64(variation(carID, trackID+5, 2))*0.125)
	}
	fmt.Fprintf(&sb, "# %s\n", footer)
	return sb.String()
}

// comboInfo is the resolved car+track pair both generators build from.
type comboInfo struct {
	carName, carCat                             string
	trackName, configName, trackCat, trackLabel string
}

// lookupCombo resolves the car (required) and track (optional; trackID 0 means a
// track-agnostic baseline) from the catalog.
func (s *Service) lookupCombo(ctx context.Context, carID, trackID int) (comboInfo, error) {
	var info comboInfo
	err := s.db.QueryRowContext(ctx,
		`SELECT car_name, category FROM cars WHERE car_id = ?`, carID).Scan(&info.carName, &info.carCat)
	if errors.Is(err, sql.ErrNoRows) {
		return comboInfo{}, ErrInvalid
	}
	if err != nil {
		return comboInfo{}, err
	}

	info.trackName, info.trackCat = "Any Track", "road"
	if trackID > 0 {
		err = s.db.QueryRowContext(ctx,
			`SELECT track_name, COALESCE(config_name, ''), category FROM tracks WHERE track_id = ?`,
			trackID).Scan(&info.trackName, &info.configName, &info.trackCat)
		if errors.Is(err, sql.ErrNoRows) {
			return comboInfo{}, ErrInvalid
		}
		if err != nil {
			return comboInfo{}, err
		}
	}

	info.trackLabel = info.trackName
	if info.configName != "" {
		info.trackLabel += " (" + info.configName + ")"
	}
	return info, nil
}

// Generate builds a single balanced baseline setup for the car+track combo.
func (s *Service) Generate(ctx context.Context, carID, trackID int) (Generated, error) {
	info, err := s.lookupCombo(ctx, carID, trackID)
	if err != nil {
		return Generated{}, err
	}

	b := baselineFor(info.carCat)
	tw := tweakFor(info.trackCat, info.trackName, info.configName)
	c := computeSetup(b, tw, delta{}, carID, trackID, 0)
	footer := b.notes + "\n# Track: " + tw.label
	data := render(info.carName, info.trackLabel, info.trackCat, c, carID, trackID, footer)

	return Generated{
		Name:  fmt.Sprintf("%s @ %s — baseline", shorten(info.carName, 40), shorten(info.trackName, 40)),
		Notes: fmt.Sprintf("Generated baseline (%s; %s). Tune from here — it is a starting point, not a hotlap file.", info.carCat, tw.label),
		Data:  data,
	}, nil
}

// GeneratePack builds the full 2×4 matrix of setups for the combo: each
// skill (safe/pro) × session (endurance/race/qual/rain) pairing, deterministic
// and ready to save. The single-baseline Generate is the "safe... race"-ish
// centre of this space; the pack spreads out from it.
func (s *Service) GeneratePack(ctx context.Context, carID, trackID int) ([]GeneratedVariant, error) {
	info, err := s.lookupCombo(ctx, carID, trackID)
	if err != nil {
		return nil, err
	}

	b := baselineFor(info.carCat)
	tw := tweakFor(info.trackCat, info.trackName, info.configName)

	out := make([]GeneratedVariant, 0, len(packOrder))
	for i, v := range packOrder {
		d := addDelta(skillDeltas[v.skill], sessionDeltas[v.session])
		// salt spreads the deterministic jitter so no two variants collide.
		c := computeSetup(b, tw, d, carID, trackID, (i+1)*7)

		km, sm := skillMeta[v.skill], sessionMeta[v.session]
		label := km.name + " · " + sm.name
		footer := fmt.Sprintf("%s — %s; %s\n# Car: %s\n# Track: %s", label, km.note, sm.note, b.notes, tw.label)
		data := render(info.carName, info.trackLabel, info.trackCat, c, carID, trackID, footer)

		out = append(out, GeneratedVariant{
			Generated: Generated{
				Name:  fmt.Sprintf("%s @ %s — %s", shorten(info.carName, 28), shorten(info.trackName, 28), label),
				Notes: fmt.Sprintf("%s setup (%s). %s Track: %s. Tune from here.", label, info.carCat, sm.note, tw.label),
				Data:  data,
			},
			Skill:   v.skill,
			Session: v.session,
			Label:   label,
		})
	}
	return out, nil
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

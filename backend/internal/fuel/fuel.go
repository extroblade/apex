// Package fuel computes a race fuel & pit-stop strategy from a few inputs.
//
// It has no dependencies and touches no database or network — it's pure
// calculation, which makes it a good first Go package to read and to test.
package fuel

import (
	"errors"
	"math"
)

// RaceType selects how the race length is expressed. Defining a named string
// type (instead of a bare string) documents intent and lets us group the valid
// values as constants below.
type RaceType string

const (
	ByLaps RaceType = "laps"
	ByTime RaceType = "time"
)

// Strategy selects how laps are distributed across stints.
type Strategy string

const (
	// Balanced splits the race into near-equal stints (e.g. 21 laps on a
	// 20-lap tank -> 11 + 10, not 20 + 1).
	Balanced Strategy = "balanced"
	// Undercut shortens the first stint to pit early and gain track position
	// on fresh tires.
	Undercut Strategy = "undercut"
	// Overcut lengthens the first stint to stay out and jump rivals who
	// pitted before you.
	Overcut Strategy = "overcut"
)

// WindowUnit expresses pit-window bounds in laps or race minutes.
type WindowUnit string

const (
	UnitLaps    WindowUnit = "laps"
	UnitMinutes WindowUnit = "minutes"
)

// PitWindow bounds when the i-th mandatory stop may happen. Zero values mean
// "no bound on this side".
type PitWindow struct {
	From float64    `json:"from"`
	To   float64    `json:"to"`
	Unit WindowUnit `json:"unit"`
}

// Rules are optional race regulations the plan must satisfy.
type Rules struct {
	// MandatoryStops is the minimum number of pit stops, even if fuel would
	// last the distance.
	MandatoryStops int `json:"mandatoryStops"`
	// Windows optionally constrains each stop (Windows[0] -> first stop, ...).
	Windows []PitWindow `json:"windows"`
}

// Request is the fuel-plan input. The `json:"..."` struct tags tell Go's JSON
// encoder/decoder what field names to use on the wire (camelCase for JS).
type Request struct {
	RaceType     RaceType `json:"raceType"`
	RaceLaps     int      `json:"raceLaps"`     // used when RaceType == ByLaps
	RaceMinutes  float64  `json:"raceMinutes"`  // used when RaceType == ByTime
	LapTimeSec   float64  `json:"lapTimeSec"`   // average lap time, seconds
	FuelPerLap   float64  `json:"fuelPerLap"`   // liters burned per lap
	TankCapacity float64  `json:"tankCapacity"` // liters
	ExtraLaps    int      `json:"extraLaps"`    // safety margin, in laps of fuel
	Strategy     Strategy `json:"strategy"`     // defaults to Balanced
	Rules        Rules    `json:"rules"`        // optional regulations
}

// Stint is one run between pit stops.
type Stint struct {
	Index int     `json:"index"`
	Laps  int     `json:"laps"`
	Fuel  float64 `json:"fuel"`     // liters to load for this stint
	PitOn int     `json:"pitOnLap"` // cumulative lap of the pit at stint end; 0 = race end
}

// Plan is the computed strategy returned to the client.
type Plan struct {
	TotalLaps      int     `json:"totalLaps"`
	LapsPerStint   int     `json:"lapsPerStint"`   // max laps on a full tank
	FuelNeeded     float64 `json:"fuelNeeded"`     // liters, no margin
	FuelWithMargin float64 `json:"fuelWithMargin"` // liters incl. ExtraLaps
	StartFuel      float64 `json:"startFuel"`      // liters to start the race with
	PitStops       int     `json:"pitStops"`
	Stints         []Stint `json:"stints"`
}

// Sentinel errors. Callers can compare against these with errors.Is to react
// to specific failures; here we mostly surface the message to the API client.
var (
	ErrFuelPerLap      = errors.New("fuelPerLap must be greater than 0")
	ErrTankCapacity    = errors.New("tankCapacity must be at least fuelPerLap")
	ErrRaceLength      = errors.New("race length is invalid; set raceLaps or raceMinutes")
	ErrLapTime         = errors.New("lapTimeSec must be greater than 0 for a timed race")
	ErrWindowLapTime   = errors.New("lapTimeSec is required when a pit window is set in minutes")
	ErrRulesInfeasible = errors.New("the pit-stop rules cannot be satisfied for this race")
	ErrTooManyStops    = errors.New("mandatoryStops is unreasonably high for this race length")
)

// Calculate builds a fuel plan from the request, or returns an error describing
// why the input is invalid. Returning (value, error) is the standard Go pattern.
func Calculate(r Request) (Plan, error) {
	if r.FuelPerLap <= 0 {
		return Plan{}, ErrFuelPerLap
	}
	if r.TankCapacity < r.FuelPerLap {
		return Plan{}, ErrTankCapacity
	}

	total, err := totalLaps(r)
	if err != nil {
		return Plan{}, err
	}

	lapsPerStint := int(math.Floor(r.TankCapacity / r.FuelPerLap))

	bounds, err := windowBounds(r, total)
	if err != nil {
		return Plan{}, err
	}
	stints, err := buildStints(total, lapsPerStint, r.FuelPerLap, r.TankCapacity,
		r.ExtraLaps, r.Strategy, r.Rules.MandatoryStops, bounds)
	if err != nil {
		return Plan{}, err
	}

	return Plan{
		TotalLaps:      total,
		LapsPerStint:   lapsPerStint,
		FuelNeeded:     round2(float64(total) * r.FuelPerLap),
		FuelWithMargin: round2(float64(total+r.ExtraLaps) * r.FuelPerLap),
		StartFuel:      stints[0].Fuel,
		PitStops:       len(stints) - 1,
		Stints:         stints,
	}, nil
}

// totalLaps resolves the race distance to a lap count.
func totalLaps(r Request) (int, error) {
	switch r.RaceType {
	case ByLaps:
		if r.RaceLaps <= 0 {
			return 0, ErrRaceLength
		}
		return r.RaceLaps, nil
	case ByTime:
		if r.RaceMinutes <= 0 {
			return 0, ErrRaceLength
		}
		if r.LapTimeSec <= 0 {
			return 0, ErrLapTime
		}
		// The leader takes the checkered on the lap *after* the timer hits zero,
		// so we floor the division and add one.
		return int(math.Floor(r.RaceMinutes*60/r.LapTimeSec)) + 1, nil
	default:
		return 0, ErrRaceLength
	}
}

// lapBounds is a pit window converted to cumulative-lap bounds.
type lapBounds struct {
	lo, hi int // hi == 0 means unbounded above
}

// windowBounds converts the rules' windows to lap bounds. Minute-based windows
// need the average lap time to translate race time into laps.
func windowBounds(r Request, total int) ([]lapBounds, error) {
	bounds := make([]lapBounds, len(r.Rules.Windows))
	for i, w := range r.Rules.Windows {
		b := lapBounds{}
		switch w.Unit {
		case UnitMinutes:
			if r.LapTimeSec <= 0 {
				return nil, ErrWindowLapTime
			}
			if w.From > 0 {
				b.lo = int(math.Ceil(w.From * 60 / r.LapTimeSec))
			}
			if w.To > 0 {
				b.hi = int(math.Floor(w.To * 60 / r.LapTimeSec))
			}
		default: // laps
			b.lo = int(w.From)
			b.hi = int(w.To)
		}
		if b.hi != 0 && (b.hi < b.lo || b.hi >= total) {
			// A pit can't be later than the second-to-last lap.
			if b.hi >= total {
				b.hi = total - 1
			}
			if b.hi < b.lo {
				return nil, ErrRulesInfeasible
			}
		}
		bounds[i] = b
	}
	return bounds, nil
}

// buildStints splits the race into stints per the chosen strategy and rules.
// It never produces lopsided splits like 20+1 — Balanced yields near-equal
// stints (21 laps / 20-lap tank -> 11+10); Undercut/Overcut bias the first pit
// earlier/later by ~20% of a stint. Mandatory stops can force extra stints, and
// each stop is clamped into its pit window (or ErrRulesInfeasible).
func buildStints(total, lapsPerStint int, fuelPerLap, tank float64, extraLaps int,
	strategy Strategy, mandatoryStops int, windows []lapBounds,
) ([]Stint, error) {
	if mandatoryStops < 0 {
		mandatoryStops = 0
	}
	fuelStops := (total+lapsPerStint-1)/lapsPerStint - 1
	stops := max(fuelStops, mandatoryStops)
	if stops+1 > total { // every stint needs at least one lap
		return nil, ErrTooManyStops
	}
	n := stops + 1

	// Cumulative pit laps, near-equally spaced, with the strategy bias applied
	// to the first stop.
	c := make([]int, stops)
	for i := range c {
		c[i] = int(math.Round(float64(total) * float64(i+1) / float64(n)))
	}
	if stops > 0 {
		bias := max(1, (total/n)/5)
		switch strategy {
		case Undercut:
			c[0] -= bias
		case Overcut:
			c[0] += bias
		}
	}

	// Forward pass: monotonic, within tank range from the previous pit, and
	// inside each stop's window.
	prev := 0
	for i := range c {
		lo, hi := prev+1, prev+lapsPerStint
		if i < len(windows) {
			if windows[i].lo > lo {
				lo = windows[i].lo
			}
			if windows[i].hi != 0 && windows[i].hi < hi {
				hi = windows[i].hi
			}
		}
		if lo > hi {
			return nil, ErrRulesInfeasible
		}
		c[i] = min(max(c[i], lo), hi)
		prev = c[i]
	}

	// Backward pass: make sure every remaining stint can reach the finish
	// within tank range, raising pits as needed (but never past their window).
	upper := total
	for i := stops - 1; i >= 0; i-- {
		need := upper - lapsPerStint // pit i must be at/after this lap
		if c[i] < need {
			c[i] = need
		}
		hi := upper - 1
		if i < len(windows) && windows[i].hi != 0 && windows[i].hi < hi {
			hi = windows[i].hi
		}
		if c[i] > hi {
			return nil, ErrRulesInfeasible
		}
		upper = c[i]
	}
	if stops > 0 && c[0] > lapsPerStint {
		return nil, ErrRulesInfeasible
	}

	// Materialize stints from the pit positions.
	stints := make([]Stint, 0, n)
	prev = 0
	for i := 0; i < n; i++ {
		end := total
		pitOn := 0
		if i < stops {
			end = c[i]
			pitOn = c[i]
		}
		laps := end - prev
		fuel := float64(laps) * fuelPerLap
		if i == n-1 { // final stint carries the safety margin
			fuel += float64(extraLaps) * fuelPerLap
		}
		fuel = math.Min(fuel, tank)
		stints = append(stints, Stint{Index: i + 1, Laps: laps, Fuel: roundUp1(fuel), PitOn: pitOn})
		prev = end
	}
	return stints, nil
}

// round2 rounds to 2 decimals for display. roundUp1 rounds *up* to 0.1 L so a
// loadout is never short.
func round2(v float64) float64   { return math.Round(v*100) / 100 }
func roundUp1(v float64) float64 { return math.Ceil(v*10) / 10 }

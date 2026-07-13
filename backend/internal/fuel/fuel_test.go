package fuel

import (
	"errors"
	"testing"
)

// Table-driven tests are the standard Go style: define a slice of cases, then
// loop over them. t.Run gives each case its own named subtest in the output.
func TestCalculate(t *testing.T) {
	cases := []struct {
		name      string
		in        Request
		wantLaps  int
		wantStops int
		wantErr   error
	}{
		{
			name: "lap race, two stints",
			in: Request{
				RaceType:     ByLaps,
				RaceLaps:     30,
				FuelPerLap:   2.5,
				TankCapacity: 50, // 20 laps per tank -> 30 laps needs 2 stints
			},
			wantLaps:  30,
			wantStops: 1,
		},
		{
			name: "timed race rounds up a lap",
			in: Request{
				RaceType:     ByTime,
				RaceMinutes:  20,
				LapTimeSec:   90, // 20*60/90 = 13.3 -> floor 13, +1 = 14 laps
				FuelPerLap:   3,
				TankCapacity: 60,
			},
			wantLaps:  14,
			wantStops: 0,
		},
		{
			name:    "zero fuel per lap is rejected",
			in:      Request{RaceType: ByLaps, RaceLaps: 10, FuelPerLap: 0, TankCapacity: 50},
			wantErr: ErrFuelPerLap,
		},
		{
			name:    "tank smaller than a lap is rejected",
			in:      Request{RaceType: ByLaps, RaceLaps: 10, FuelPerLap: 5, TankCapacity: 4},
			wantErr: ErrTankCapacity,
		},
		{
			name:    "missing race length is rejected",
			in:      Request{RaceType: ByLaps, FuelPerLap: 2, TankCapacity: 50},
			wantErr: ErrRaceLength,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := Calculate(tc.in)

			if tc.wantErr != nil {
				if !errors.Is(err, tc.wantErr) {
					t.Fatalf("want error %v, got %v", tc.wantErr, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got.TotalLaps != tc.wantLaps {
				t.Errorf("TotalLaps: want %d, got %d", tc.wantLaps, got.TotalLaps)
			}
			if got.PitStops != tc.wantStops {
				t.Errorf("PitStops: want %d, got %d", tc.wantStops, got.PitStops)
			}
			// The stints should always add up to the total race distance.
			sum := 0
			for _, s := range got.Stints {
				sum += s.Laps
			}
			if sum != got.TotalLaps {
				t.Errorf("stint laps sum to %d, want %d", sum, got.TotalLaps)
			}
		})
	}
}

// The user-reported case: 21 laps on a 20-lap tank must not be 20+1.
func TestBalancedStints(t *testing.T) {
	got, err := Calculate(Request{
		RaceType: ByLaps, RaceLaps: 21, FuelPerLap: 2.5, TankCapacity: 50,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(got.Stints) != 2 || got.Stints[0].Laps != 11 || got.Stints[1].Laps != 10 {
		t.Fatalf("want balanced 11+10, got %+v", got.Stints)
	}
}

// Mandatory stops force extra stints even when fuel would last the distance.
func TestMandatoryStops(t *testing.T) {
	got, err := Calculate(Request{
		RaceType: ByLaps, RaceLaps: 30, FuelPerLap: 1, TankCapacity: 100, // no fuel stop needed
		Rules: Rules{MandatoryStops: 2},
	})
	if err != nil {
		t.Fatal(err)
	}
	if got.PitStops != 2 || len(got.Stints) != 3 {
		t.Fatalf("want 2 stops / 3 stints, got %d/%d", got.PitStops, len(got.Stints))
	}
	if got.Stints[0].PitOn == 0 || got.Stints[1].PitOn <= got.Stints[0].PitOn {
		t.Errorf("pit laps should be increasing: %+v", got.Stints)
	}
}

// The user's example: 60-minute race, 90s laps, 2 mandatory stops — first pit
// after minute 5, second between lap 15 and lap 45.
func TestPitWindows(t *testing.T) {
	got, err := Calculate(Request{
		RaceType: ByTime, RaceMinutes: 60, LapTimeSec: 90,
		FuelPerLap: 1, TankCapacity: 100,
		Rules: Rules{
			MandatoryStops: 2,
			Windows: []PitWindow{
				{From: 5, Unit: UnitMinutes},
				{From: 15, To: 45, Unit: UnitLaps},
			},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	// 60*60/90 = 40 -> +1 = 41 laps; 5 min at 90s = lap 4 earliest.
	if got.TotalLaps != 41 {
		t.Fatalf("total laps: want 41, got %d", got.TotalLaps)
	}
	first, second := got.Stints[0].PitOn, got.Stints[1].PitOn
	if first < 4 {
		t.Errorf("first pit must be at/after lap 4 (minute 5), got %d", first)
	}
	if second < 15 || second > 45 {
		t.Errorf("second pit must be within laps 15..45, got %d", second)
	}
}

// A window can also drag a fuel-driven stop earlier than the balanced split.
func TestWindowClampsFuelStop(t *testing.T) {
	got, err := Calculate(Request{
		RaceType: ByLaps, RaceLaps: 30, FuelPerLap: 2.5, TankCapacity: 50, // needs 1 stop
		Rules: Rules{Windows: []PitWindow{{To: 12, Unit: UnitLaps}}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if got.Stints[0].PitOn > 12 {
		t.Errorf("pit must respect the lap-12 deadline, got %d", got.Stints[0].PitOn)
	}
}

func TestInfeasibleRules(t *testing.T) {
	// Second stint would need 28 laps on a 20-lap tank: impossible.
	_, err := Calculate(Request{
		RaceType: ByLaps, RaceLaps: 30, FuelPerLap: 2.5, TankCapacity: 50,
		Rules: Rules{Windows: []PitWindow{{To: 2, Unit: UnitLaps}}},
	})
	if !errors.Is(err, ErrRulesInfeasible) {
		t.Errorf("want ErrRulesInfeasible, got %v", err)
	}

	// Minute-based window without a lap time is rejected.
	_, err = Calculate(Request{
		RaceType: ByLaps, RaceLaps: 30, FuelPerLap: 2.5, TankCapacity: 50,
		Rules: Rules{MandatoryStops: 1, Windows: []PitWindow{{From: 5, Unit: UnitMinutes}}},
	})
	if !errors.Is(err, ErrWindowLapTime) {
		t.Errorf("want ErrWindowLapTime, got %v", err)
	}
}

func TestStrategies(t *testing.T) {
	base := Request{RaceType: ByLaps, RaceLaps: 30, FuelPerLap: 2.5, TankCapacity: 50}

	run := func(s Strategy) Plan {
		t.Helper()
		r := base
		r.Strategy = s
		p, err := Calculate(r)
		if err != nil {
			t.Fatal(err)
		}
		sum := 0
		for _, st := range p.Stints {
			sum += st.Laps
			if st.Laps < 1 || st.Laps > p.LapsPerStint {
				t.Fatalf("%s: stint out of bounds: %+v", s, st)
			}
		}
		if sum != p.TotalLaps {
			t.Fatalf("%s: laps sum %d != %d", s, sum, p.TotalLaps)
		}
		return p
	}

	balanced := run(Balanced) // 30 laps, 20/tank -> 15+15
	under := run(Undercut)
	over := run(Overcut)

	if balanced.Stints[0].Laps != 15 {
		t.Errorf("balanced first stint: want 15, got %d", balanced.Stints[0].Laps)
	}
	if under.Stints[0].Laps >= balanced.Stints[0].Laps {
		t.Errorf("undercut first stint (%d) should be shorter than balanced (%d)",
			under.Stints[0].Laps, balanced.Stints[0].Laps)
	}
	if over.Stints[0].Laps <= balanced.Stints[0].Laps {
		t.Errorf("overcut first stint (%d) should be longer than balanced (%d)",
			over.Stints[0].Laps, balanced.Stints[0].Laps)
	}
}

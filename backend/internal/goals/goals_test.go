package goals

import (
	"context"
	"errors"
	"testing"
)

func TestCreate_Validation(t *testing.T) {
	s := New(nil)
	cases := []struct {
		name string
		in   Input
	}{
		{"empty title", Input{Title: "  ", Target: 5}},
		{"negative target", Input{Title: "Win races", Target: -1}},
		{"negative current", Input{Title: "Win races", Current: -3}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if _, err := s.Create(context.Background(), 1, c.in); !errors.Is(err, ErrInvalid) {
				t.Fatalf("Create(%+v) err = %v, want ErrInvalid", c.in, err)
			}
		})
	}
}

func TestProgressAndAutoDone(t *testing.T) {
	tests := []struct {
		current, target float64
		done            bool
		want            float64
	}{
		{0, 10, false, 0},
		{5, 10, false, 0.5},
		{10, 10, false, 1},
		{15, 10, false, 1}, // clamped
		{3, 0, false, 0},   // no target
		{3, 10, true, 1},   // manually done
	}
	for _, tt := range tests {
		if got := progress(tt.current, tt.target, tt.done); got != tt.want {
			t.Errorf("progress(%v,%v,%v) = %v, want %v", tt.current, tt.target, tt.done, got, tt.want)
		}
	}

	if !autoDone(Input{Target: 10, Current: 10}) {
		t.Error("autoDone should be true when current reaches target")
	}
	if autoDone(Input{Target: 10, Current: 4}) {
		t.Error("autoDone should be false below target")
	}
	no := false
	if autoDone(Input{Target: 10, Current: 10, Done: &no}) {
		t.Error("explicit done=false should win over derived done")
	}
}

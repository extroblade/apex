package setups

import (
	"context"
	"errors"
	"testing"
)

// Create rejects invalid input before touching the DB, so a nil-DB service is
// enough to exercise the validation branch.
func TestCreate_Validation(t *testing.T) {
	s := New(nil)
	cases := []struct {
		name string
		in   Input
	}{
		{"empty name", Input{CarID: 1, Data: "x", Name: "  "}},
		{"no car", Input{CarID: 0, Data: "x", Name: "Baseline"}},
		{"no data", Input{CarID: 1, Data: "", Name: "Baseline"}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if _, err := s.Create(context.Background(), 1, c.in); !errors.Is(err, ErrInvalid) {
				t.Fatalf("Create(%+v) err = %v, want ErrInvalid", c.in, err)
			}
		})
	}
}

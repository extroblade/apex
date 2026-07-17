package main

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

type fakeStore struct {
	items []Item
	err   error
}

func (f *fakeStore) List(context.Context) ([]Item, error) { return f.items, f.err }

func TestSplitPlacements(t *testing.T) {
	cases := map[string][]string{
		"side,bottom":   {"side", "bottom"},
		"side, bottom ": {"side", "bottom"},
		"side":          {"side"},
		"":              {},
		",,":            {},
	}
	for in, want := range cases {
		got := splitPlacements(in)
		if len(got) != len(want) {
			t.Errorf("splitPlacements(%q) = %v, want %v", in, got, want)
			continue
		}
		for i := range want {
			if got[i] != want[i] {
				t.Errorf("splitPlacements(%q) = %v, want %v", in, got, want)
				break
			}
		}
	}
}

func TestNavHandler(t *testing.T) {
	items := []Item{
		{Key: "home", LabelKey: "nav.home", Href: "/", Icon: "home",
			Placements: []string{"side", "bottom"}, Order: 10},
		{Key: "drivers", LabelKey: "nav.drivers", Href: "/drivers", Icon: "users",
			Placements: []string{"side"}, Order: 70, RequiresAuth: true, FeatureFlag: "iracing_oauth"},
	}

	rec := httptest.NewRecorder()
	navHandler(&fakeStore{items: items})(rec, httptest.NewRequest(http.MethodGet, "/api/nav", nil))

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	var body struct {
		Items []Item `json:"items"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(body.Items) != 2 {
		t.Fatalf("got %d items, want 2", len(body.Items))
	}
	// The gating metadata must survive to the client, which does the filtering.
	if !body.Items[1].RequiresAuth || body.Items[1].FeatureFlag != "iracing_oauth" {
		t.Errorf("gating metadata lost: %+v", body.Items[1])
	}
	// An ungated item must not carry a flag (omitempty keeps the payload clean).
	if body.Items[0].FeatureFlag != "" {
		t.Errorf("unexpected flag on home: %q", body.Items[0].FeatureFlag)
	}
}

func TestNavHandlerStoreError(t *testing.T) {
	rec := httptest.NewRecorder()
	navHandler(&fakeStore{err: errors.New("db down")})(rec, httptest.NewRequest(http.MethodGet, "/api/nav", nil))
	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", rec.Code)
	}
}

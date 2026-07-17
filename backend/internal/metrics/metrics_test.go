package metrics

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
)

func TestMiddlewareRecordsRequests(t *testing.T) {
	r := chi.NewRouter()
	r.Use(Middleware)
	r.Get("/api/things/{id}", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})

	before := testutil.ToFloat64(reqTotal.WithLabelValues("GET", "/api/things/{id}", "204"))

	// Two requests to different ids share the same route-pattern label.
	for _, id := range []string{"1", "2"} {
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/things/"+id, nil))
		if rec.Code != http.StatusNoContent {
			t.Fatalf("status = %d, want 204", rec.Code)
		}
	}

	after := testutil.ToFloat64(reqTotal.WithLabelValues("GET", "/api/things/{id}", "204"))
	if after-before != 2 {
		t.Errorf("counter delta = %v, want 2 (route pattern should collapse ids)", after-before)
	}
}

func TestCountCreatesAndIncrements(t *testing.T) {
	labels := prometheus.Labels{"kind": "pack"}
	Count("apex_test_events_total", "Test events.", labels)
	Count("apex_test_events_total", "Test events.", labels)

	got := testutil.ToFloat64(counters["apex_test_events_total"].With(labels))
	if got != 2 {
		t.Errorf("count = %v, want 2", got)
	}
}

func TestHandlerServesExposition(t *testing.T) {
	Count("apex_exposition_probe_total", "Probe.", prometheus.Labels{"x": "y"})

	rec := httptest.NewRecorder()
	Handler().ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/metrics", nil))

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "apex_exposition_probe_total") {
		t.Errorf("exposition missing the domain counter")
	}
}

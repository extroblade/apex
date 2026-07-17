// Package metrics provides Prometheus instrumentation: an HTTP middleware that
// records request counts + latencies, the /metrics exposition handler, and a
// generic domain counter (Count) that mirrors the frontend's
// counter('event')(labels) — one universal entry point for any metric.
package metrics

import (
	"net/http"
	"sort"
	"strconv"
	"sync"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	reqTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "apex_http_requests_total",
		Help: "HTTP requests by method, route and status.",
	}, []string{"method", "route", "status"})

	reqDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "apex_http_request_duration_seconds",
		Help:    "HTTP request duration in seconds by method and route.",
		Buckets: prometheus.DefBuckets,
	}, []string{"method", "route"})
)

// Handler serves the Prometheus exposition format (mount at GET /metrics).
func Handler() http.Handler { return promhttp.Handler() }

// Middleware records each request's count and duration. It labels by the chi
// ROUTE PATTERN (e.g. "/api/setups/{id}"), not the raw path, so per-id URLs
// don't explode metric cardinality.
func Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rec := &statusRecorder{ResponseWriter: w, status: http.StatusOK}

		next.ServeHTTP(rec, r)

		route := chi.RouteContext(r.Context()).RoutePattern()
		if route == "" {
			route = "unmatched"
		}
		reqDuration.WithLabelValues(r.Method, route).Observe(time.Since(start).Seconds())
		reqTotal.WithLabelValues(r.Method, route, strconv.Itoa(rec.status)).Inc()
	})
}

type statusRecorder struct {
	http.ResponseWriter
	status  int
	written bool
}

func (s *statusRecorder) WriteHeader(code int) {
	if !s.written {
		s.status = code
		s.written = true
	}
	s.ResponseWriter.WriteHeader(code)
}

func (s *statusRecorder) Write(b []byte) (int, error) {
	s.written = true // a bare Write implies 200
	return s.ResponseWriter.Write(b)
}

var (
	mu       sync.Mutex
	counters = map[string]*prometheus.CounterVec{}
)

// Count increments a named domain counter, lazily creating it on first use. The
// label KEYS fix the metric's label set, so a given name must always be called
// with the same keys (a Prometheus requirement). This is the generic
// "any metric" entry point, e.g.:
//
//	metrics.Count("apex_setups_generated_total", "Setups generated.",
//		prometheus.Labels{"kind": "pack"})
func Count(name, help string, labels prometheus.Labels) {
	keys := make([]string, 0, len(labels))
	for k := range labels {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	mu.Lock()
	cv, ok := counters[name]
	if !ok {
		cv = promauto.NewCounterVec(prometheus.CounterOpts{Name: name, Help: help}, keys)
		counters[name] = cv
	}
	mu.Unlock()

	cv.With(labels).Inc()
}

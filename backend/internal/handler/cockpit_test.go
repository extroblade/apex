package handler

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"

	"apex/internal/cache"
	"apex/internal/features"
)

func devReq(method, path, cookie string) *http.Request {
	r := httptest.NewRequest(method, path, strings.NewReader(`{"enabled":true}`))
	if cookie != "" {
		r.AddCookie(&http.Cookie{Name: "developer", Value: cookie})
	}
	return r
}

func TestAllFeaturesGating(t *testing.T) {
	cases := []struct {
		name   string
		key    string // DEVELOPER_KEY
		cookie string
		want   int
	}{
		{"disabled when key empty", "", "anything", http.StatusNotFound},
		{"404 without cookie", "secret", "", http.StatusNotFound},
		{"404 on wrong cookie", "secret", "nope", http.StatusNotFound},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// Features present so only the cookie gate is exercised.
			db, _, _ := sqlmock.New()
			defer db.Close()
			h := &Handler{Features: features.NewService(db), DeveloperKey: tc.key}
			rec := httptest.NewRecorder()
			h.AllFeatures(rec, devReq(http.MethodGet, "/api/features/all", tc.cookie))
			if rec.Code != tc.want {
				t.Errorf("status = %d, want %d", rec.Code, tc.want)
			}
		})
	}
}

func TestAllFeaturesAuthorized(t *testing.T) {
	db, mock, _ := sqlmock.New()
	defer db.Close()
	mock.ExpectQuery("SELECT flag_key, enabled FROM feature_flags").WillReturnRows(
		sqlmock.NewRows([]string{"flag_key", "enabled"}).AddRow("cockpit", true))

	h := &Handler{Features: features.NewService(db), DeveloperKey: "secret"}
	rec := httptest.NewRecorder()
	h.AllFeatures(rec, devReq(http.MethodGet, "/api/features/all", "secret"))

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), `"cockpit":true`) {
		t.Errorf("body = %s", rec.Body.String())
	}
}

func TestHealthCockpit(t *testing.T) {
	db, mock, _ := sqlmock.New(sqlmock.MonitorPingsOption(true))
	defer db.Close()
	mock.ExpectPing()

	// Redis disabled (blank addr) → redisEnabled=false, no ping attempted.
	h := &Handler{DB: db, Cache: cache.New(""), DeveloperKey: "secret"}
	rec := httptest.NewRecorder()
	h.HealthCockpit(rec, devReq(http.MethodGet, "/api/health/cockpit", "secret"))

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	body := rec.Body.String()
	if !strings.Contains(body, `"db":true`) || !strings.Contains(body, `"redisEnabled":false`) {
		t.Errorf("body = %s", body)
	}
}

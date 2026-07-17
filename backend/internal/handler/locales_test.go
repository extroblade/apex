package handler

import (
	"database/sql"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/go-chi/chi/v5"

	"apex/internal/locales"
)

func TestListLocalesHandler(t *testing.T) {
	db, mock, _ := sqlmock.New()
	defer db.Close()
	mock.ExpectQuery("SELECT code, name FROM locales").WillReturnRows(
		sqlmock.NewRows([]string{"code", "name"}).AddRow("en", "English").AddRow("ru", "Русский"))

	h := &Handler{Locales: locales.New(db)}
	rec := httptest.NewRecorder()
	h.ListLocales(rec, httptest.NewRequest(http.MethodGet, "/api/locales", nil))

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), `"code":"ru"`) {
		t.Errorf("body missing ru: %s", rec.Body.String())
	}
}

func TestGetLocaleHandler(t *testing.T) {
	db, mock, _ := sqlmock.New()
	defer db.Close()
	mock.ExpectQuery("SELECT bundle FROM locales").WithArgs("ru").
		WillReturnRows(sqlmock.NewRows([]string{"bundle"}).AddRow(`{"nav":{"home":"Главная"}}`))
	mock.ExpectQuery("SELECT bundle FROM locales").WithArgs("zz").
		WillReturnError(sql.ErrNoRows)

	r := chi.NewRouter()
	r.Get("/api/locales/{code}", (&Handler{Locales: locales.New(db)}).GetLocale)

	// Found → the raw bundle JSON.
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/locales/ru", nil))
	if rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), "Главная") {
		t.Fatalf("found: status=%d body=%s", rec.Code, rec.Body.String())
	}
	if ct := rec.Header().Get("Content-Type"); ct != "application/json" {
		t.Errorf("content-type = %q", ct)
	}

	// Unknown → 404.
	rec = httptest.NewRecorder()
	r.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/locales/zz", nil))
	if rec.Code != http.StatusNotFound {
		t.Errorf("unknown locale status = %d, want 404", rec.Code)
	}
}

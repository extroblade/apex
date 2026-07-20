package server

import (
	"database/sql"
	"testing"

	"apex/internal/config"

	"github.com/DATA-DOG/go-sqlmock"
)

// TestNew_DoesNotPanic guards against the chi panic "all middlewares must be
// defined before routes on a mux" — a regression introduced when middleware
// registration was placed after r.Handle/r.Route. It builds the router against
// a sqlmock DB and asserts the call returns without panicking.
func TestNew_DoesNotPanic(t *testing.T) {
	mockDB, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer mockDB.Close()

	cfg := &config.Config{
		CookieSecure:    false,
		CORSOrigin:      "*",
		DeveloperKey:    "3",
		AuthRateLimit:   0,
		EncryptionKey:   "",
		IRacingClientID: "",
	}

	// server.New only stores the *sql.DB; it doesn't run queries at construction
	// time, so no sqlmock expectations are needed.
	var db *sql.DB = mockDB
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("server.New panicked: %v", r)
		}
	}()
	_ = New(cfg, db)
}

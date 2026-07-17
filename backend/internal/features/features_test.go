package features

import (
	"context"
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

func flagRows() *sqlmock.Rows {
	return sqlmock.NewRows([]string{"flag_key", "enabled"}).
		AddRow("cockpit", false).
		AddRow("iracing_oauth", true)
}

func TestAllAndEnabled(t *testing.T) {
	db, mock, _ := sqlmock.New()
	defer db.Close()
	mock.ExpectQuery("SELECT flag_key, enabled FROM feature_flags").WillReturnRows(flagRows())

	s := NewService(db)
	all, err := s.All(context.Background())
	if err != nil {
		t.Fatalf("All: %v", err)
	}
	if !all["iracing_oauth"] || all["cockpit"] {
		t.Errorf("All = %+v", all)
	}
	// A second read is served from the in-process cache — no new DB query, so the
	// single expectation above suffices.
	if s.Enabled(context.Background(), "iracing_oauth") != true {
		t.Error("Enabled(iracing_oauth) = false, want true")
	}
	if s.Enabled(context.Background(), "missing") != false {
		t.Error("unknown flag should be off")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Error(err)
	}
}

func TestSetSuccess(t *testing.T) {
	db, mock, _ := sqlmock.New()
	defer db.Close()
	mock.ExpectExec("UPDATE feature_flags SET enabled").
		WithArgs(true, "cockpit").
		WillReturnResult(sqlmock.NewResult(0, 1))

	if err := NewService(db).Set(context.Background(), "cockpit", true); err != nil {
		t.Fatalf("Set: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Error(err)
	}
}

func TestSetUnknownFlag(t *testing.T) {
	db, mock, _ := sqlmock.New()
	defer db.Close()
	mock.ExpectExec("UPDATE feature_flags SET enabled").
		WithArgs(true, "nope").
		WillReturnResult(sqlmock.NewResult(0, 0)) // no rows affected

	err := NewService(db).Set(context.Background(), "nope", true)
	if !errors.Is(err, ErrFlagNotFound) {
		t.Errorf("err = %v, want ErrFlagNotFound", err)
	}
}

func TestInvalidateForcesReload(t *testing.T) {
	db, mock, _ := sqlmock.New()
	defer db.Close()
	// Two loads expected: one before Invalidate, one after.
	mock.ExpectQuery("SELECT flag_key, enabled FROM feature_flags").WillReturnRows(flagRows())
	mock.ExpectQuery("SELECT flag_key, enabled FROM feature_flags").WillReturnRows(flagRows())

	s := NewService(db)
	if _, err := s.All(context.Background()); err != nil {
		t.Fatal(err)
	}
	s.Invalidate()
	if _, err := s.All(context.Background()); err != nil {
		t.Fatal(err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Error(err)
	}
}

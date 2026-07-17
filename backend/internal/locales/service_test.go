package locales

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestList(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	mock.ExpectQuery("SELECT code, name FROM locales").WillReturnRows(
		sqlmock.NewRows([]string{"code", "name"}).
			AddRow("en", "English").
			AddRow("ru", "Русский"))

	got, err := New(db).List(context.Background())
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(got) != 2 || got[0].Code != "en" || got[1].Name != "Русский" {
		t.Errorf("List = %+v", got)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Error(err)
	}
}

func TestBundleFound(t *testing.T) {
	db, mock, _ := sqlmock.New()
	defer db.Close()

	mock.ExpectQuery("SELECT bundle FROM locales").WithArgs("ru").
		WillReturnRows(sqlmock.NewRows([]string{"bundle"}).AddRow(`{"nav":{"home":"Главная"}}`))

	got, err := New(db).Bundle(context.Background(), "ru")
	if err != nil {
		t.Fatalf("Bundle: %v", err)
	}
	if got != `{"nav":{"home":"Главная"}}` {
		t.Errorf("Bundle = %q", got)
	}
}

func TestBundleNotFound(t *testing.T) {
	db, mock, _ := sqlmock.New()
	defer db.Close()

	mock.ExpectQuery("SELECT bundle FROM locales").WithArgs("zz").
		WillReturnError(sql.ErrNoRows)

	_, err := New(db).Bundle(context.Background(), "zz")
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("err = %v, want ErrNotFound", err)
	}
}

func TestSeedUpsertsBuiltins(t *testing.T) {
	db, mock, _ := sqlmock.New()
	defer db.Close()

	// One upsert per built-in (en, ru), reading the embedded bundle each time.
	for range builtins {
		mock.ExpectExec("INSERT INTO locales").WillReturnResult(sqlmock.NewResult(1, 1))
	}

	if err := Seed(context.Background(), db); err != nil {
		t.Fatalf("Seed: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Error(err)
	}
}

package racing

import (
	"context"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"

	"apex/internal/iracing"
)

// Free content is owned by everyone — the catalog read should report it as owned
// even when the user has no ownership row. This is what makes the garage show
// free cars/tracks as already-checked (read-only) instead of "missing".
func TestCars_FreeReportedAsOwned(t *testing.T) {
	db, mock, _ := sqlmock.New()
	defer db.Close()
	mock.ExpectQuery("SELECT c.car_id, c.car_name, c.category, c.description, c.is_free,").
		WithArgs(int64(7)).
		WillReturnRows(sqlmock.NewRows([]string{"car_id", "car_name", "category", "description", "is_free", "owned"}).
			AddRow(1, "Spec Racer Ford", "road", "", true, true).   // free, no ownership row → owned
			AddRow(2, "Skip Barber Formula 2000", "road", "", false, false). // paid, no ownership → not owned
			AddRow(3, "Formula Vee", "road", "", false, true))      // paid, owned → owned

	s := NewService(db, nil, func() APIClient { return &fakeClient{} }, iracing.OAuthConfig{})
	items, err := s.Cars(context.Background(), 7)
	if err != nil {
		t.Fatalf("Cars: %v", err)
	}
	if len(items) != 3 {
		t.Fatalf("want 3 items, got %d", len(items))
	}
	// Free car must be owned even without an ownership row.
	if !items[0].Owned {
		t.Errorf("free car: Owned = false, want true (is_free should imply owned)")
	}
	if !items[0].Free {
		t.Errorf("free car: Free = false, want true")
	}
	// Paid car without ownership must not be owned.
	if items[1].Owned {
		t.Errorf("paid unowned car: Owned = true, want false")
	}
	// Paid car with ownership must be owned.
	if !items[2].Owned {
		t.Errorf("owned paid car: Owned = false, want true")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Error(err)
	}
}

func TestTracks_FreeReportedAsOwned(t *testing.T) {
	db, mock, _ := sqlmock.New()
	defer db.Close()
	mock.ExpectQuery("SELECT t.track_id, t.track_name, t.config_name, t.category, t.description, t.is_free,").
		WithArgs(int64(7)).
		WillReturnRows(sqlmock.NewRows([]string{"track_id", "track_name", "config_name", "category", "description", "is_free", "owned"}).
			AddRow(1, "Summit Point", "Summit Point Raceway", "road", "", true, true).
			AddRow(2, "Watkins Glen", "Boot", "road", "", false, false))

	s := NewService(db, nil, func() APIClient { return &fakeClient{} }, iracing.OAuthConfig{})
	items, err := s.Tracks(context.Background(), 7)
	if err != nil {
		t.Fatalf("Tracks: %v", err)
	}
	if len(items) != 2 {
		t.Fatalf("want 2 items, got %d", len(items))
	}
	if !items[0].Owned || !items[0].Free {
		t.Errorf("free track: Owned=%v Free=%v, want both true", items[0].Owned, items[0].Free)
	}
	if items[1].Owned {
		t.Errorf("paid unowned track: Owned = true, want false")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Error(err)
	}
}

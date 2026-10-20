package postgres

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"fmt"
	"io"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/ashadashraf/ride-hail-app/internal/domain/ride"
	"github.com/google/uuid"
)

var rideRepoTestDriverID atomic.Uint64

type rideRepoTestState struct {
	execArgs []driver.NamedValue
	row      []driver.Value
}

type rideRepoTestDriver struct{ state *rideRepoTestState }

func (d rideRepoTestDriver) Open(string) (driver.Conn, error) {
	return rideRepoTestConn{state: d.state}, nil
}

type rideRepoTestConn struct{ state *rideRepoTestState }

func (c rideRepoTestConn) Prepare(string) (driver.Stmt, error) {
	return nil, errors.New("Prepare not supported")
}
func (c rideRepoTestConn) Close() error              { return nil }
func (c rideRepoTestConn) Begin() (driver.Tx, error) { return rideRepoTestTx{}, nil }
func (c rideRepoTestConn) BeginTx(context.Context, driver.TxOptions) (driver.Tx, error) {
	return rideRepoTestTx{}, nil
}
func (c rideRepoTestConn) ExecContext(_ context.Context, _ string, args []driver.NamedValue) (driver.Result, error) {
	c.state.execArgs = append([]driver.NamedValue(nil), args...)
	return driver.RowsAffected(1), nil
}
func (c rideRepoTestConn) QueryContext(_ context.Context, query string, _ []driver.NamedValue) (driver.Rows, error) {
	columns := []string{"id", "rider_id", "driver_id", "pickup_latitude", "pickup_longitude", "status", "created_at", "updated_at"}
	row := c.state.row
	if strings.Contains(query, "version") {
		columns = []string{"id", "rider_id", "driver_id", "pickup_latitude", "pickup_longitude", "status", "version", "created_at", "updated_at"}
		row = append(append([]driver.Value(nil), row[:6]...), append([]driver.Value{int64(3)}, row[6:]...)...)
	}
	return &rideRepoTestRows{columns: columns, rows: [][]driver.Value{row}}, nil
}

type rideRepoTestTx struct{}

func (rideRepoTestTx) Commit() error   { return nil }
func (rideRepoTestTx) Rollback() error { return nil }

type rideRepoTestRows struct {
	columns []string
	rows    [][]driver.Value
	index   int
}

func (r *rideRepoTestRows) Columns() []string { return r.columns }
func (r *rideRepoTestRows) Close() error      { return nil }
func (r *rideRepoTestRows) Next(dest []driver.Value) error {
	if r.index >= len(r.rows) {
		return io.EOF
	}
	copy(dest, r.rows[r.index])
	r.index++
	return nil
}

func newRideRepositoryTestDB(t *testing.T, state *rideRepoTestState) *sql.DB {
	t.Helper()
	name := fmt.Sprintf("ride-repository-test-%d", rideRepoTestDriverID.Add(1))
	sql.Register(name, rideRepoTestDriver{state: state})
	db, err := sql.Open(name, "")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })
	return db
}

func TestRideRepositoryPersistsAndLoadsPickupCoordinates(t *testing.T) {
	rideID, riderID := uuid.New(), uuid.New()
	createdAt := time.Date(2026, 10, 3, 12, 0, 0, 0, time.UTC)
	state := &rideRepoTestState{row: []driver.Value{
		rideID.String(), riderID.String(), nil, float64(0), float64(0), string(ride.StatusRequested), createdAt, createdAt,
	}}
	repository := NewRideRepository(newRideRepositoryTestDB(t, state))
	entity := &ride.Ride{
		ID: rideID, RiderID: riderID, Status: ride.StatusRequested,
		PickupLatitude: 0, PickupLongitude: 0, CreatedAt: createdAt, UpdatedAt: createdAt,
	}

	if err := repository.Save(context.Background(), entity); err != nil {
		t.Fatalf("Save() error = %v", err)
	}
	if len(state.execArgs) != 8 || state.execArgs[3].Value != float64(0) || state.execArgs[4].Value != float64(0) {
		t.Fatalf("Save() args do not contain pickup coordinates in expected positions: %#v", state.execArgs)
	}

	loaded, err := repository.GetByID(context.Background(), rideID)
	if err != nil {
		t.Fatalf("GetByID() error = %v", err)
	}
	if loaded.PickupLatitude != 0 || loaded.PickupLongitude != 0 {
		t.Fatalf("loaded pickup = (%v, %v), want (0, 0)", loaded.PickupLatitude, loaded.PickupLongitude)
	}

	tx, err := repository.db.BeginTx(context.Background(), nil)
	if err != nil {
		t.Fatal(err)
	}
	if err := repository.SaveTx(context.Background(), tx, entity); err != nil {
		t.Fatalf("SaveTx() error = %v", err)
	}
	if err := tx.Rollback(); err != nil {
		t.Fatal(err)
	}
	if len(state.execArgs) != 9 || state.execArgs[3].Value != float64(0) || state.execArgs[4].Value != float64(0) {
		t.Fatalf("SaveTx() args do not contain pickup coordinates in expected positions: %#v", state.execArgs)
	}
}

func TestRideRepositoryActiveRideQueryLoadsPickupCoordinates(t *testing.T) {
	rideID, riderID := uuid.New(), uuid.New()
	createdAt := time.Date(2026, 10, 3, 12, 0, 0, 0, time.UTC)
	state := &rideRepoTestState{row: []driver.Value{
		rideID.String(), riderID.String(), nil, float64(9.9), float64(76.2), string(ride.StatusMatching), createdAt, createdAt,
	}}
	repository := NewRideRepository(newRideRepositoryTestDB(t, state))
	rides, err := repository.GetActiveRides(context.Background())
	if err != nil {
		t.Fatalf("GetActiveRides() error = %v", err)
	}
	if len(rides) != 1 || rides[0].PickupLatitude != 9.9 || rides[0].PickupLongitude != 76.2 {
		t.Fatalf("GetActiveRides() = %#v, want one ride with pickup (9.9, 76.2)", rides)
	}
}

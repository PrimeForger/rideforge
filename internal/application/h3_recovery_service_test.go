package application

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"testing"

	"github.com/ashadashraf/ride-hail-app/internal/domain/driver"
	"github.com/ashadashraf/ride-hail-app/internal/domain/ride"
	"github.com/ashadashraf/ride-hail-app/internal/ports"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// --- Mock Driver Repository ---

type mockDriverRepo struct {
	availableDrivers []*driver.Driver
	err              error
}

func (m *mockDriverRepo) GetAvailableDrivers(ctx context.Context) ([]*driver.Driver, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.availableDrivers, nil
}

func (m *mockDriverRepo) GetAvailableDriversExcludingTx(ctx context.Context, tx *sql.Tx, rideID uuid.UUID) ([]*driver.Driver, error) {
	return nil, nil
}
func (m *mockDriverRepo) GetEligibleDriversTx(ctx context.Context, tx *sql.Tx, rideID uuid.UUID, driverIDs []uuid.UUID) ([]*driver.Driver, error) {
	return nil, nil
}
func (m *mockDriverRepo) GetByID(ctx context.Context, id uuid.UUID) (*driver.Driver, error) {
	return nil, nil
}
func (m *mockDriverRepo) GetByIDTx(ctx context.Context, tx *sql.Tx, id uuid.UUID) (*driver.Driver, error) {
	return nil, nil
}
func (m *mockDriverRepo) GetByIDs(ctx context.Context, ids []uuid.UUID) ([]*driver.Driver, error) {
	return nil, nil
}
func (m *mockDriverRepo) Save(ctx context.Context, d *driver.Driver) error { return nil }
func (m *mockDriverRepo) SaveTx(ctx context.Context, tx *sql.Tx, d *driver.Driver) error {
	return nil
}
func (m *mockDriverRepo) InsertRideOfferTx(ctx context.Context, tx *sql.Tx, rideID, driverID uuid.UUID, attempt int) error {
	return nil
}
func (m *mockDriverRepo) MarkDriverAcceptedTx(ctx context.Context, tx *sql.Tx, rideID, driverID uuid.UUID) error {
	return nil
}
func (m *mockDriverRepo) MarkDriverRejectedTx(ctx context.Context, tx *sql.Tx, rideID, driverID uuid.UUID) error {
	return nil
}
func (m *mockDriverRepo) MarkDriverTimeoutTx(ctx context.Context, tx *sql.Tx, rideID, driverID uuid.UUID) error {
	return nil
}
func (m *mockDriverRepo) CountRideAttemptsTx(ctx context.Context, tx *sql.Tx, rideID uuid.UUID) (int, error) {
	return 0, nil
}
func (m *mockDriverRepo) UpdateOfferStatusTx(ctx context.Context, tx *sql.Tx, rideID, driverID uuid.UUID, status ride.OfferStatus) error {
	return nil
}
func (m *mockDriverRepo) GetActiveOfferDriversTx(ctx context.Context, tx *sql.Tx, rideID uuid.UUID, excludeDriverID uuid.UUID) ([]uuid.UUID, error) {
	return nil, nil
}
func (m *mockDriverRepo) ExpireOtherOffersTx(ctx context.Context, tx *sql.Tx, rideID uuid.UUID, acceptedDriverID uuid.UUID) error {
	return nil
}
func (m *mockDriverRepo) MarkDriverBusyTx(ctx context.Context, tx *sql.Tx, driverID uuid.UUID) error {
	return nil
}

// --- Mock H3 Cell Calculator ---

type mockH3Calc struct {
	err error
}

func (m *mockH3Calc) CellForLocation(lat, lng float64) (string, error) {
	if m.err != nil {
		return "", m.err
	}
	return fmt.Sprintf("cell_%.2f_%.2f", lat, lng), nil
}

// --- Mock H3 Driver Index ---

type mockH3Index struct {
	updatedCells map[uuid.UUID]string
	removedIDs   []uuid.UUID
	updateErr    error
	removeErr    error
}

func newMockH3Index() *mockH3Index {
	return &mockH3Index{
		updatedCells: make(map[uuid.UUID]string),
	}
}

func (m *mockH3Index) UpdateDriverCell(ctx context.Context, driverID uuid.UUID, newCell string) (ports.CellUpdateResult, error) {
	if m.updateErr != nil {
		return ports.CellUpdateResult{}, m.updateErr
	}
	oldCell := m.updatedCells[driverID]
	status := ports.DriverCellAdded
	if oldCell != "" {
		if oldCell == newCell {
			status = ports.DriverCellUnchanged
		} else {
			status = ports.DriverCellMoved
		}
	}
	m.updatedCells[driverID] = newCell
	return ports.CellUpdateResult{
		Status:  status,
		OldCell: oldCell,
		NewCell: newCell,
	}, nil
}

func (m *mockH3Index) RemoveDriver(ctx context.Context, driverID uuid.UUID) (ports.DriverRemoveResult, error) {
	if m.removeErr != nil {
		return ports.DriverRemoveResult{}, m.removeErr
	}
	oldCell := m.updatedCells[driverID]
	delete(m.updatedCells, driverID)
	m.removedIDs = append(m.removedIDs, driverID)
	return ports.DriverRemoveResult{
		Removed: oldCell != "",
		OldCell: oldCell,
	}, nil
}

// --- Mock Geo Indexer ---

type mockGeoIndex struct {
	locations map[uuid.UUID][2]float64
	removed   []uuid.UUID
	updateErr error
	removeErr error
}

func newMockGeoIndex() *mockGeoIndex {
	return &mockGeoIndex{
		locations: make(map[uuid.UUID][2]float64),
	}
}

func (m *mockGeoIndex) UpdateDriverLocation(ctx context.Context, driverID uuid.UUID, lat, lng float64) error {
	if m.updateErr != nil {
		return m.updateErr
	}
	m.locations[driverID] = [2]float64{lat, lng}
	return nil
}

func (m *mockGeoIndex) RemoveDriver(ctx context.Context, driverID uuid.UUID) error {
	if m.removeErr != nil {
		return m.removeErr
	}
	delete(m.locations, driverID)
	m.removed = append(m.removed, driverID)
	return nil
}

// --- Mock Realtime Driver Cache ---

type mockDriverCache struct {
	onlineIDs     []uuid.UUID
	onlineIDsErr  error
	cachedDrivers map[uuid.UUID]*driver.Driver
	loadErr       error
}

func (m *mockDriverCache) GetOnlineDriverIDs(ctx context.Context) ([]uuid.UUID, error) {
	if m.onlineIDsErr != nil {
		return nil, m.onlineIDsErr
	}
	return m.onlineIDs, nil
}

func (m *mockDriverCache) LoadDrivers(ctx context.Context, driverIDs []uuid.UUID) ([]*driver.Driver, error) {
	if m.loadErr != nil {
		return nil, m.loadErr
	}
	res := make([]*driver.Driver, 0, len(driverIDs))
	for _, id := range driverIDs {
		if d, ok := m.cachedDrivers[id]; ok {
			res = append(res, d)
		}
	}
	return res, nil
}

// --- Unit Tests ---

func TestH3RecoveryService_EmptyState(t *testing.T) {
	ctx := context.Background()
	repo := &mockDriverRepo{availableDrivers: []*driver.Driver{}}
	h3Calc := &mockH3Calc{}
	h3Idx := newMockH3Index()
	geoIdx := newMockGeoIndex()
	log := zap.NewNop()

	svc := NewH3RecoveryService(repo, h3Calc, h3Idx, geoIdx, nil, log)

	res, err := svc.RebuildIndex(ctx)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if res.TotalDrivers != 0 || res.RestoredDrivers != 0 || res.SkippedDrivers != 0 {
		t.Errorf("expected all zero counts, got %+v", res)
	}
}

func TestH3RecoveryService_SingleDriver(t *testing.T) {
	ctx := context.Background()
	dID := uuid.New()
	d := &driver.Driver{
		ID:     dID,
		Status: driver.StatusOnline,
		Lat:    12.9716,
		Lng:    77.5946,
	}

	repo := &mockDriverRepo{availableDrivers: []*driver.Driver{d}}
	h3Calc := &mockH3Calc{}
	h3Idx := newMockH3Index()
	geoIdx := newMockGeoIndex()
	log := zap.NewNop()

	svc := NewH3RecoveryService(repo, h3Calc, h3Idx, geoIdx, nil, log)

	res, err := svc.RebuildIndex(ctx)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if res.RestoredDrivers != 1 || res.SkippedDrivers != 0 {
		t.Errorf("expected 1 restored driver, got %+v", res)
	}

	expectedCell := "cell_12.97_77.59"
	if cell, ok := h3Idx.updatedCells[dID]; !ok || cell != expectedCell {
		t.Errorf("expected H3 cell %s, got %s", expectedCell, cell)
	}

	if loc, ok := geoIdx.locations[dID]; !ok || loc[0] != 12.9716 || loc[1] != 77.5946 {
		t.Errorf("expected Geo location [12.9716, 77.5946], got %v", loc)
	}
}

func TestH3RecoveryService_MultipleDrivers(t *testing.T) {
	ctx := context.Background()
	d1 := &driver.Driver{ID: uuid.New(), Status: driver.StatusOnline, Lat: 12.9, Lng: 77.5}
	d2 := &driver.Driver{ID: uuid.New(), Status: driver.StatusOnline, Lat: 13.0, Lng: 77.6}
	d3 := &driver.Driver{ID: uuid.New(), Status: driver.StatusOnline, Lat: 13.1, Lng: 77.7}

	repo := &mockDriverRepo{availableDrivers: []*driver.Driver{d1, d2, d3}}
	h3Calc := &mockH3Calc{}
	h3Idx := newMockH3Index()
	geoIdx := newMockGeoIndex()
	log := zap.NewNop()

	svc := NewH3RecoveryService(repo, h3Calc, h3Idx, geoIdx, nil, log)

	res, err := svc.RebuildIndex(ctx)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if res.RestoredDrivers != 3 || res.SkippedDrivers != 0 {
		t.Errorf("expected 3 restored drivers, got %+v", res)
	}

	if len(h3Idx.updatedCells) != 3 {
		t.Errorf("expected 3 H3 cell updates, got %d", len(h3Idx.updatedCells))
	}
}

func TestH3RecoveryService_RedisOnlyDriverIgnored(t *testing.T) {
	ctx := context.Background()
	pgDriverID := uuid.New()
	redisOnlyDriverID := uuid.New()

	pgDriver := &driver.Driver{ID: pgDriverID, Status: driver.StatusOnline, Lat: 12.9, Lng: 77.5}
	redisOnlyDriver := &driver.Driver{ID: redisOnlyDriverID, Status: driver.StatusOnline, Lat: 13.0, Lng: 77.6}

	repo := &mockDriverRepo{availableDrivers: []*driver.Driver{pgDriver}}
	h3Calc := &mockH3Calc{}
	h3Idx := newMockH3Index()
	geoIdx := newMockGeoIndex()
	log := zap.NewNop()

	cache := &mockDriverCache{
		onlineIDs: []uuid.UUID{pgDriverID, redisOnlyDriverID},
		cachedDrivers: map[uuid.UUID]*driver.Driver{
			pgDriverID:        pgDriver,
			redisOnlyDriverID: redisOnlyDriver,
		},
	}

	svc := NewH3RecoveryService(repo, h3Calc, h3Idx, geoIdx, cache, log)

	res, err := svc.RebuildIndex(ctx)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if res.RestoredDrivers != 1 {
		t.Errorf("expected exactly 1 restored driver (only PostgreSQL driver), got %d", res.RestoredDrivers)
	}

	if _, found := h3Idx.updatedCells[pgDriverID]; !found {
		t.Errorf("expected PostgreSQL driver %s to be restored", pgDriverID)
	}

	if _, found := h3Idx.updatedCells[redisOnlyDriverID]; found {
		t.Errorf("expected Redis-only driver %s NOT to be restored", redisOnlyDriverID)
	}
}

func TestH3RecoveryService_DriverCacheFailureContinues(t *testing.T) {
	ctx := context.Background()
	dID := uuid.New()
	pgDriver := &driver.Driver{ID: dID, Status: driver.StatusOnline, Lat: 12.9, Lng: 77.5}

	repo := &mockDriverRepo{availableDrivers: []*driver.Driver{pgDriver}}
	h3Calc := &mockH3Calc{}
	h3Idx := newMockH3Index()
	geoIdx := newMockGeoIndex()
	log := zap.NewNop()

	cache := &mockDriverCache{
		onlineIDsErr: errors.New("redis connection timeout"),
	}

	svc := NewH3RecoveryService(repo, h3Calc, h3Idx, geoIdx, cache, log)

	res, err := svc.RebuildIndex(ctx)
	if err != nil {
		t.Fatalf("expected recovery to succeed despite DriverCache failure, got error: %v", err)
	}

	if res.RestoredDrivers != 1 {
		t.Errorf("expected 1 restored driver using PostgreSQL coordinates, got %d", res.RestoredDrivers)
	}

	if cell, ok := h3Idx.updatedCells[dID]; !ok || cell != "cell_12.90_77.50" {
		t.Errorf("expected driver restored with PostgreSQL cell cell_12.90_77.50, got %s", cell)
	}
}

func TestH3RecoveryService_InvalidAndOfflineDrivers(t *testing.T) {
	ctx := context.Background()
	dValid := &driver.Driver{ID: uuid.New(), Status: driver.StatusOnline, Lat: 12.9, Lng: 77.5}
	dOfflineID := uuid.New()
	dOffline := &driver.Driver{ID: dOfflineID, Status: driver.StatusOffline, Lat: 12.9, Lng: 77.5}
	dZeroCoordsID := uuid.New()
	dZeroCoords := &driver.Driver{ID: dZeroCoordsID, Status: driver.StatusOnline, Lat: 0, Lng: 0}
	dBadCoordsID := uuid.New()
	dBadCoords := &driver.Driver{ID: dBadCoordsID, Status: driver.StatusOnline, Lat: 100.0, Lng: 77.5}

	repo := &mockDriverRepo{availableDrivers: []*driver.Driver{dValid, dOffline, dZeroCoords, dBadCoords}}
	h3Calc := &mockH3Calc{}
	h3Idx := newMockH3Index()
	geoIdx := newMockGeoIndex()
	log := zap.NewNop()

	// Pre-populate index to verify explicit removal of invalid/offline drivers
	h3Idx.updatedCells[dOfflineID] = "old_cell"
	geoIdx.locations[dOfflineID] = [2]float64{12.9, 77.5}

	svc := NewH3RecoveryService(repo, h3Calc, h3Idx, geoIdx, nil, log)

	res, err := svc.RebuildIndex(ctx)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if res.RestoredDrivers != 1 {
		t.Errorf("expected 1 restored driver, got %d", res.RestoredDrivers)
	}
	if res.SkippedDrivers != 3 {
		t.Errorf("expected 3 skipped drivers, got %d", res.SkippedDrivers)
	}

	// Verify dOffline was explicitly removed from derived indexes
	if _, found := h3Idx.updatedCells[dOfflineID]; found {
		t.Errorf("expected offline driver %s to be removed from H3 index", dOfflineID)
	}
	if _, found := geoIdx.locations[dOfflineID]; found {
		t.Errorf("expected offline driver %s to be removed from Geo index", dOfflineID)
	}
}

func TestH3RecoveryService_Idempotency(t *testing.T) {
	ctx := context.Background()
	dID := uuid.New()
	d := &driver.Driver{ID: dID, Status: driver.StatusOnline, Lat: 12.9716, Lng: 77.5946}

	repo := &mockDriverRepo{availableDrivers: []*driver.Driver{d}}
	h3Calc := &mockH3Calc{}
	h3Idx := newMockH3Index()
	geoIdx := newMockGeoIndex()
	log := zap.NewNop()

	svc := NewH3RecoveryService(repo, h3Calc, h3Idx, geoIdx, nil, log)

	// First execution
	res1, err := svc.RebuildIndex(ctx)
	if err != nil {
		t.Fatalf("first recovery failed: %v", err)
	}

	// Second execution
	res2, err := svc.RebuildIndex(ctx)
	if err != nil {
		t.Fatalf("second recovery failed: %v", err)
	}

	if res1.RestoredDrivers != res2.RestoredDrivers {
		t.Errorf("expected same restored count across runs, run 1=%d, run 2=%d", res1.RestoredDrivers, res2.RestoredDrivers)
	}

	if res2.CellUpdates != 0 {
		t.Errorf("expected 0 cell updates on second run due to unchanged cell, got %d", res2.CellUpdates)
	}
}

func TestH3RecoveryService_DriverMovement(t *testing.T) {
	ctx := context.Background()
	dID := uuid.New()

	repo := &mockDriverRepo{}
	h3Calc := &mockH3Calc{}
	h3Idx := newMockH3Index()
	geoIdx := newMockGeoIndex()
	log := zap.NewNop()

	svc := NewH3RecoveryService(repo, h3Calc, h3Idx, geoIdx, nil, log)

	// Run 1: Driver at Location A
	repo.availableDrivers = []*driver.Driver{{ID: dID, Status: driver.StatusOnline, Lat: 12.9, Lng: 77.5}}
	_, _ = svc.RebuildIndex(ctx)
	if h3Idx.updatedCells[dID] != "cell_12.90_77.50" {
		t.Fatalf("unexpected initial cell: %s", h3Idx.updatedCells[dID])
	}

	// Run 2: Driver moved to Location B
	repo.availableDrivers = []*driver.Driver{{ID: dID, Status: driver.StatusOnline, Lat: 13.5, Lng: 78.0}}
	res2, err := svc.RebuildIndex(ctx)
	if err != nil {
		t.Fatalf("second recovery failed: %v", err)
	}

	if res2.CellUpdates != 1 {
		t.Errorf("expected 1 cell update for moved driver, got %d", res2.CellUpdates)
	}
	if h3Idx.updatedCells[dID] != "cell_13.50_78.00" {
		t.Errorf("expected moved cell cell_13.50_78.00, got %s", h3Idx.updatedCells[dID])
	}
}

func TestH3RecoveryService_StaleDriverCleanup(t *testing.T) {
	ctx := context.Background()
	activeID := uuid.New()
	staleID := uuid.New()

	activeDriver := &driver.Driver{ID: activeID, Status: driver.StatusOnline, Lat: 12.9, Lng: 77.5}

	repo := &mockDriverRepo{availableDrivers: []*driver.Driver{activeDriver}}
	h3Calc := &mockH3Calc{}
	h3Idx := newMockH3Index()
	geoIdx := newMockGeoIndex()
	log := zap.NewNop()

	// Pretend staleID is currently in Redis H3/Geo index
	h3Idx.updatedCells[staleID] = "cell_old"
	geoIdx.locations[staleID] = [2]float64{10.0, 10.0}

	cache := &mockDriverCache{
		onlineIDs: []uuid.UUID{activeID, staleID},
		cachedDrivers: map[uuid.UUID]*driver.Driver{
			activeID: activeDriver,
		},
	}

	svc := NewH3RecoveryService(repo, h3Calc, h3Idx, geoIdx, cache, log)

	_, err := svc.RebuildIndex(ctx)
	if err != nil {
		t.Fatalf("recovery failed: %v", err)
	}

	// Verify staleID was removed
	if _, found := h3Idx.updatedCells[staleID]; found {
		t.Errorf("expected stale driver %s to be removed from H3 index", staleID)
	}
	if _, found := geoIdx.locations[staleID]; found {
		t.Errorf("expected stale driver %s to be removed from Geo index", staleID)
	}
}

func TestH3RecoveryService_StaleH3CleanupFailure(t *testing.T) {
	ctx := context.Background()
	activeID := uuid.New()
	staleID := uuid.New()

	activeDriver := &driver.Driver{ID: activeID, Status: driver.StatusOnline, Lat: 12.9, Lng: 77.5}

	repo := &mockDriverRepo{availableDrivers: []*driver.Driver{activeDriver}}
	h3Calc := &mockH3Calc{}
	h3Idx := newMockH3Index()
	h3Idx.removeErr = errors.New("h3 remove operation failed")
	geoIdx := newMockGeoIndex()
	log := zap.NewNop()

	cache := &mockDriverCache{
		onlineIDs: []uuid.UUID{activeID, staleID},
	}

	svc := NewH3RecoveryService(repo, h3Calc, h3Idx, geoIdx, cache, log)

	_, err := svc.RebuildIndex(ctx)
	if err == nil {
		t.Fatal("expected error when stale H3 cleanup fails, got nil")
	}
}

func TestH3RecoveryService_StaleGeoCleanupFailure(t *testing.T) {
	ctx := context.Background()
	activeID := uuid.New()
	staleID := uuid.New()

	activeDriver := &driver.Driver{ID: activeID, Status: driver.StatusOnline, Lat: 12.9, Lng: 77.5}

	repo := &mockDriverRepo{availableDrivers: []*driver.Driver{activeDriver}}
	h3Calc := &mockH3Calc{}
	h3Idx := newMockH3Index()
	geoIdx := newMockGeoIndex()
	geoIdx.removeErr = errors.New("geo remove operation failed")
	log := zap.NewNop()

	cache := &mockDriverCache{
		onlineIDs: []uuid.UUID{activeID, staleID},
	}

	svc := NewH3RecoveryService(repo, h3Calc, h3Idx, geoIdx, cache, log)

	_, err := svc.RebuildIndex(ctx)
	if err == nil {
		t.Fatal("expected error when stale Geo cleanup fails, got nil")
	}
}

func TestH3RecoveryService_RepositoryFailure(t *testing.T) {
	ctx := context.Background()
	repoErr := errors.New("database connection failed")
	repo := &mockDriverRepo{err: repoErr}
	h3Calc := &mockH3Calc{}
	h3Idx := newMockH3Index()
	geoIdx := newMockGeoIndex()
	log := zap.NewNop()

	svc := NewH3RecoveryService(repo, h3Calc, h3Idx, geoIdx, nil, log)

	_, err := svc.RebuildIndex(ctx)
	if err == nil {
		t.Fatal("expected error on repository failure, got nil")
	}
	if !errors.Is(err, repoErr) {
		t.Errorf("expected error to wrap repo failure, got %v", err)
	}
}

func TestH3RecoveryService_H3IndexFailure(t *testing.T) {
	ctx := context.Background()
	d := &driver.Driver{ID: uuid.New(), Status: driver.StatusOnline, Lat: 12.9, Lng: 77.5}
	repo := &mockDriverRepo{availableDrivers: []*driver.Driver{d}}
	h3Calc := &mockH3Calc{}
	h3Idx := newMockH3Index()
	h3Idx.updateErr = errors.New("redis connection refused")
	geoIdx := newMockGeoIndex()
	log := zap.NewNop()

	svc := NewH3RecoveryService(repo, h3Calc, h3Idx, geoIdx, nil, log)

	_, err := svc.RebuildIndex(ctx)
	if err == nil {
		t.Fatal("expected error on H3 index failure, got nil")
	}
}

func TestH3RecoveryService_GeoIndexFailure(t *testing.T) {
	ctx := context.Background()
	d := &driver.Driver{ID: uuid.New(), Status: driver.StatusOnline, Lat: 12.9, Lng: 77.5}
	repo := &mockDriverRepo{availableDrivers: []*driver.Driver{d}}
	h3Calc := &mockH3Calc{}
	h3Idx := newMockH3Index()
	geoIdx := newMockGeoIndex()
	geoIdx.updateErr = errors.New("redis GEOADD timeout")
	log := zap.NewNop()

	svc := NewH3RecoveryService(repo, h3Calc, h3Idx, geoIdx, nil, log)

	_, err := svc.RebuildIndex(ctx)
	if err == nil {
		t.Fatal("expected error on Geo index failure, got nil")
	}
}

package ranking_test

import (
	"context"
	"testing"
	"time"

	"github.com/ashadashraf/ride-hail-app/internal/application/dispatch/candidate"
	"github.com/ashadashraf/ride-hail-app/internal/application/dispatch/candidate/pipeline"
	"github.com/ashadashraf/ride-hail-app/internal/application/dispatch/candidate/ranking"
	"github.com/ashadashraf/ride-hail-app/internal/application/dispatch/routing"
	"github.com/ashadashraf/ride-hail-app/internal/domain/driver"
	"github.com/google/uuid"
)

type fakeTravelCalculator struct {
	called            bool
	receivedPickupLat float64
	receivedPickupLng float64
	receivedDriver    *driver.Driver
	returnedFeatures  ranking.TravelFeatures
	returnedErr       error
}

func (f *fakeTravelCalculator) Calculate(
	ctx context.Context,
	pickupLat float64,
	pickupLng float64,
	d *driver.Driver,
) (ranking.TravelFeatures, error) {
	f.called = true
	f.receivedPickupLat = pickupLat
	f.receivedPickupLng = pickupLng
	f.receivedDriver = d
	return f.returnedFeatures, f.returnedErr
}

func TestDefaultFeatureExtractor_Extract_WithInjectedFake(t *testing.T) {
	fakeCalc := &fakeTravelCalculator{
		returnedFeatures: ranking.TravelFeatures{
			DistanceKm: 3.14,
			ETASeconds: 0,
		},
	}

	extractor := ranking.NewDefaultFeatureExtractor(fakeCalc)
	ctx := context.Background()

	pCtx := &pipeline.Context{
		RideID:    uuid.New(),
		PickupLat: 12.9716,
		PickupLng: 77.5946,
	}

	lastAssigned := time.Now().Add(-5 * time.Minute)
	d := &driver.Driver{
		ID:               uuid.New(),
		Lat:              12.9750,
		Lng:              77.5990,
		AcceptanceRate:   0.90,
		CancellationRate: 0.05,
		TimeoutRate:      0.02,
		Rating:           4.8,
		CompletedRides:   200,
		LastAssignedAt:   lastAssigned,
	}

	c := &candidate.Candidate{
		ID:     uuid.New(),
		Driver: d,
	}

	features, err := extractor.Extract(ctx, pCtx, c)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if !fakeCalc.called {
		t.Errorf("expected injected TravelCalculator.Calculate to be called")
	}
	if fakeCalc.receivedPickupLat != pCtx.PickupLat || fakeCalc.receivedPickupLng != pCtx.PickupLng {
		t.Errorf("expected pickup coordinates (%f, %f), got (%f, %f)",
			pCtx.PickupLat, pCtx.PickupLng, fakeCalc.receivedPickupLat, fakeCalc.receivedPickupLng)
	}
	if fakeCalc.receivedDriver != d {
		t.Errorf("expected driver %v to be passed to calculator, got %v", d, fakeCalc.receivedDriver)
	}
	if features.Travel.DistanceKm != 3.14 {
		t.Errorf("expected DistanceKm to be 3.14 from fake calculator, got %f", features.Travel.DistanceKm)
	}
	if features.Quality.Rating != 4.8 {
		t.Errorf("expected Rating 4.8, got %f", features.Quality.Rating)
	}
	if features.Fairness.LastAssignedAt != lastAssigned {
		t.Errorf("expected LastAssignedAt %v, got %v", lastAssigned, features.Fairness.LastAssignedAt)
	}
}

func TestDefaultFeatureExtractor_Extract_Success(t *testing.T) {
	travelCalc := ranking.NewDefaultTravelCalculator(routing.NewDisabledProvider())
	extractor := ranking.NewDefaultFeatureExtractor(travelCalc)
	ctx := context.Background()

	pCtx := &pipeline.Context{
		RideID:    uuid.New(),
		PickupLat: 12.9716,
		PickupLng: 77.5946,
	}

	lastAssigned := time.Now().Add(-10 * time.Minute)
	d := &driver.Driver{
		ID:               uuid.New(),
		Lat:              12.9750,
		Lng:              77.5990,
		AcceptanceRate:   0.95,
		CancellationRate: 0.02,
		TimeoutRate:      0.01,
		Rating:           4.9,
		CompletedRides:   150,
		LastAssignedAt:   lastAssigned,
	}

	c := &candidate.Candidate{
		ID:     uuid.New(),
		Driver: d,
	}

	features, err := extractor.Extract(ctx, pCtx, c)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if features.Travel.DistanceKm <= 0 {
		t.Errorf("expected positive DistanceKm, got %f", features.Travel.DistanceKm)
	}
	if features.Quality.AcceptanceRate != 0.95 {
		t.Errorf("expected AcceptanceRate 0.95, got %f", features.Quality.AcceptanceRate)
	}
	if features.Quality.Rating != 4.9 {
		t.Errorf("expected Rating 4.9, got %f", features.Quality.Rating)
	}
	if features.Fairness.LastAssignedAt != lastAssigned {
		t.Errorf("expected LastAssignedAt %v, got %v", lastAssigned, features.Fairness.LastAssignedAt)
	}
}

func TestDefaultFeatureExtractor_Extract_NilCandidate(t *testing.T) {
	travelCalc := ranking.NewDefaultTravelCalculator(routing.NewDisabledProvider())
	extractor := ranking.NewDefaultFeatureExtractor(travelCalc)
	ctx := context.Background()

	pCtx := &pipeline.Context{
		RideID:    uuid.New(),
		PickupLat: 12.9716,
		PickupLng: 77.5946,
	}

	_, err := extractor.Extract(ctx, pCtx, nil)
	if err == nil {
		t.Fatalf("expected error for nil candidate, got nil")
	}
	if err.Error() != "candidate is nil" {
		t.Errorf("expected 'candidate is nil' error, got '%v'", err)
	}
}

func TestDefaultFeatureExtractor_Extract_NilDriver(t *testing.T) {
	travelCalc := ranking.NewDefaultTravelCalculator(routing.NewDisabledProvider())
	extractor := ranking.NewDefaultFeatureExtractor(travelCalc)
	ctx := context.Background()

	pCtx := &pipeline.Context{
		RideID:    uuid.New(),
		PickupLat: 12.9716,
		PickupLng: 77.5946,
	}

	c := &candidate.Candidate{
		ID:     uuid.New(),
		Driver: nil,
	}

	_, err := extractor.Extract(ctx, pCtx, c)
	if err == nil {
		t.Fatalf("expected error for candidate with nil driver, got nil")
	}
	if err.Error() != "driver is nil" {
		t.Errorf("expected 'driver is nil' error, got '%v'", err)
	}
}

func TestDefaultFeatureExtractor_Extract_NilPipelineContext(t *testing.T) {
	travelCalc := ranking.NewDefaultTravelCalculator(routing.NewDisabledProvider())
	extractor := ranking.NewDefaultFeatureExtractor(travelCalc)
	ctx := context.Background()

	c := &candidate.Candidate{
		ID: uuid.New(),
		Driver: &driver.Driver{
			ID: uuid.New(),
		},
	}

	_, err := extractor.Extract(ctx, nil, c)
	if err == nil {
		t.Fatalf("expected error for nil pipeline context, got nil")
	}
	if err.Error() != "pipeline context is nil" {
		t.Errorf("expected 'pipeline context is nil' error, got '%v'", err)
	}
}

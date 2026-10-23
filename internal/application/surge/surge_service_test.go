package surge_test

import (
	"context"
	"errors"
	"testing"
	"time"

	surge "github.com/ashadashraf/ride-hail-app/internal/application/surge"
	"github.com/ashadashraf/ride-hail-app/internal/domain/occupancy"
	"github.com/ashadashraf/ride-hail-app/internal/domain/surgepolicy"
	"github.com/ashadashraf/ride-hail-app/internal/infrastructure/geo"
	"github.com/ashadashraf/ride-hail-app/internal/ports"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type localityResolverStub struct {
	id  uuid.UUID
	err error
}

func (s localityResolverStub) ResolveH3Cell(context.Context, string) (uuid.UUID, error) {
	return s.id, s.err
}

type surgeProviderStub struct {
	policy surgepolicy.Policy
	ok     bool
}

func (s surgeProviderStub) PolicyForLocality(string) (surgepolicy.Policy, bool) {
	return s.policy, s.ok
}
func (surgeProviderStub) SnapshotState() (int, time.Time, time.Time) {
	return 0, time.Time{}, time.Time{}
}

type occupancyProviderStub struct {
	value occupancy.CellOccupancy
	err   error
}

func (s occupancyProviderStub) GetCellOccupancy(context.Context, string) (occupancy.CellOccupancy, error) {
	return s.value, s.err
}
func (occupancyProviderStub) GetMultiCellOccupancy(context.Context, []string) (map[string]occupancy.CellOccupancy, error) {
	return nil, nil
}
func (occupancyProviderStub) GetRingOccupancy(context.Context, string, int) (map[string]occupancy.CellOccupancy, error) {
	return nil, nil
}
func (occupancyProviderStub) IsCellHot(context.Context, string) (bool, error) { return false, nil }
func (occupancyProviderStub) DetectHotCells(context.Context, []string) ([]occupancy.CellOccupancy, error) {
	return nil, nil
}

func TestSurgeServiceUsesValidLocalityPolicyAndOccupancy(t *testing.T) {
	now := time.Now()
	localityID := uuid.New()
	policy := surgepolicy.Policy{
		ID: uuid.New(), LocalityID: localityID, Version: 1, Status: surgepolicy.StatusActive,
		MinMultiplier: 1, MaxMultiplier: 1.5, MinimumDemand: 3,
		EffectiveFrom: now.Add(-time.Minute),
		RatioTiers:    []surgepolicy.RatioTier{{MinImbalanceRatio: 1.2, Multiplier: 1.1}, {MinImbalanceRatio: 2, Multiplier: 1.5}},
	}
	service := surge.NewSurgeService(
		geo.NewH3Service(8, 1),
		localityResolverStub{id: localityID},
		surgeProviderStub{policy: policy, ok: true},
		occupancyProviderStub{value: occupancy.CellOccupancy{DemandCount: 8, ImbalanceRatio: 2, Status: occupancy.StatusHot}},
		nil,
		zap.NewNop(),
	)
	got := service.CalculateMultiplier(context.Background(), 12.5, 74.9)
	if got.Multiplier != 1.5 || got.Reason != "applied" {
		t.Fatalf("decision = %+v, want 1.5 applied", got)
	}
}

func TestSurgeServiceFallsBackToBaseFareOnUnavailableInputs(t *testing.T) {
	now := time.Now()
	localityID := uuid.New()
	validPolicy := surgepolicy.Policy{
		ID: uuid.New(), LocalityID: localityID, Version: 1, Status: surgepolicy.StatusActive,
		MinMultiplier: 1, MaxMultiplier: 1.5, MinimumDemand: 3, EffectiveFrom: now.Add(-time.Minute),
		RatioTiers: []surgepolicy.RatioTier{{MinImbalanceRatio: 1, Multiplier: 1.1}},
	}
	expiredPolicy := validPolicy
	expiredAt := now.Add(-time.Second)
	expiredPolicy.EffectiveUntil = &expiredAt
	inactivePolicy := validPolicy
	inactivePolicy.Status = surgepolicy.StatusInactive
	invalidPolicy := validPolicy
	invalidPolicy.MaxMultiplier = .9
	cases := []struct {
		name      string
		locality  localityResolverStub
		policies  surgeProviderStub
		occupancy occupancyProviderStub
	}{
		{name: "no locality", locality: localityResolverStub{err: ports.ErrLocalityUnresolved}},
		{name: "ambiguous locality", locality: localityResolverStub{err: ports.ErrAmbiguousLocality}},
		{name: "no active policy", locality: localityResolverStub{id: localityID}},
		{name: "expired policy", locality: localityResolverStub{id: localityID}, policies: surgeProviderStub{policy: expiredPolicy, ok: true}},
		{name: "disabled policy", locality: localityResolverStub{id: localityID}, policies: surgeProviderStub{policy: inactivePolicy, ok: true}},
		{name: "invalid policy", locality: localityResolverStub{id: localityID}, policies: surgeProviderStub{policy: invalidPolicy, ok: true}},
		{name: "occupancy unavailable", locality: localityResolverStub{id: localityID}, policies: surgeProviderStub{policy: validPolicy, ok: true}, occupancy: occupancyProviderStub{err: errors.New("redis unavailable")}},
		{name: "normal occupancy", locality: localityResolverStub{id: localityID}, policies: surgeProviderStub{policy: validPolicy, ok: true}, occupancy: occupancyProviderStub{value: occupancy.CellOccupancy{Status: occupancy.StatusNormal}}},
		{name: "below policy minimum demand", locality: localityResolverStub{id: localityID}, policies: surgeProviderStub{policy: validPolicy, ok: true}, occupancy: occupancyProviderStub{value: occupancy.CellOccupancy{DemandCount: 2, ImbalanceRatio: 5, Status: occupancy.StatusHot}}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			service := surge.NewSurgeService(geo.NewH3Service(8, 1), tc.locality, tc.policies, tc.occupancy, nil, zap.NewNop())
			got := service.CalculateMultiplier(context.Background(), 12.5, 74.9)
			if got.Multiplier != surge.BaseFareMultiplier {
				t.Fatalf("decision = %+v, want base fare", got)
			}
		})
	}
}

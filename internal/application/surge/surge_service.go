package surge

import (
	"context"
	"errors"
	"time"

	"github.com/ashadashraf/ride-hail-app/internal/domain/occupancy"
	"github.com/ashadashraf/ride-hail-app/internal/domain/surgepolicy"
	"github.com/ashadashraf/ride-hail-app/internal/ports"
	"go.uber.org/zap"
)

const BaseFareMultiplier = 1.0

type SurgeDecision struct {
	Multiplier float64
	Reason     string
}

type SurgeService struct {
	h3        ports.H3CellResolver
	locality  ports.LocalityResolver
	policies  ports.SurgePolicyProvider
	occupancy ports.CellOccupancyProvider
	observer  ports.SurgeObserver
	log       *zap.Logger
	now       func() time.Time
}

func NewSurgeService(
	h3 ports.H3CellResolver,
	locality ports.LocalityResolver,
	policies ports.SurgePolicyProvider,
	occupancy ports.CellOccupancyProvider,
	observer ports.SurgeObserver,
	log *zap.Logger,
) *SurgeService {
	return &SurgeService{h3: h3, locality: locality, policies: policies, occupancy: occupancy, observer: observer, log: log, now: time.Now}
}

// CalculateMultiplier is fail-safe: invalid, unresolved, disabled, stale, or
// unavailable surge inputs produce the base multiplier without failing quotes.
func (s *SurgeService) CalculateMultiplier(ctx context.Context, latitude, longitude float64) SurgeDecision {
	return s.calculateMultiplier(ctx, latitude, longitude)
}

func (s *SurgeService) calculateMultiplier(ctx context.Context, latitude, longitude float64) SurgeDecision {
	cell, err := s.h3.CellForLocation(latitude, longitude)
	if err != nil {
		return s.base("invalid_pickup", err)
	}
	localityID, err := s.locality.ResolveH3Cell(ctx, cell)
	if err != nil {
		reason := "locality_lookup_failed"
		loggedErr := err
		if errors.Is(err, ports.ErrLocalityUnresolved) {
			reason, loggedErr = "no_locality", nil
		}
		if errors.Is(err, ports.ErrAmbiguousLocality) {
			reason = "locality_ambiguous"
		}
		return s.base(reason, loggedErr)
	}
	policy, ok := s.policies.PolicyForLocality(localityID.String())
	if !ok {
		return s.base("policy_unavailable", nil)
	}
	now := s.now()
	if !policy.ValidAt(now) {
		return s.base("policy_invalid_or_inactive", surgepolicy.ErrInvalidPolicy)
	}
	cellOccupancy, err := s.occupancy.GetCellOccupancy(ctx, cell)
	if err != nil {
		return s.base("occupancy_unavailable", err)
	}
	if cellOccupancy.Status != occupancy.StatusWarm && cellOccupancy.Status != occupancy.StatusHot {
		return s.base("occupancy_normal", nil)
	}
	multiplier, err := policy.MultiplierFor(cellOccupancy.DemandCount, cellOccupancy.ImbalanceRatio, now)
	if err != nil {
		return s.base("calculation_invalid", err)
	}
	s.observe("applied")
	return SurgeDecision{Multiplier: multiplier, Reason: "applied"}
}

func (s *SurgeService) base(reason string, err error) SurgeDecision {
	s.observe(reason)
	if err != nil && s.log != nil {
		s.log.Warn("surge calculation fell back to base fare", zap.String("reason", reason), zap.Error(err))
	}
	return SurgeDecision{Multiplier: BaseFareMultiplier, Reason: reason}
}

func (s *SurgeService) observe(reason string) {
	if s.observer != nil {
		s.observer.ObserveSurgeDecision(reason)
	}
}

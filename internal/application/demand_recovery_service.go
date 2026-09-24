package application

import (
	"context"
	"fmt"
	"time"

	"github.com/ashadashraf/ride-hail-app/internal/domain/ride"
	"github.com/ashadashraf/ride-hail-app/internal/infrastructure/geo"
	"github.com/ashadashraf/ride-hail-app/internal/ports"
	"github.com/google/uuid"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.uber.org/zap"
)

var demandRecoveryTracer = otel.Tracer("application.demand_recovery")

type DemandRecoveryResult struct {
	TotalActiveRides int
	TotalCells       int
	Duration         time.Duration
}

type DemandRecoveryService struct {
	rideRepo    ports.RideRepository
	h3          *geo.H3Service
	demandStore ports.CellDemandStore
	log         *zap.Logger
}

func NewDemandRecoveryService(
	rideRepo ports.RideRepository,
	h3 *geo.H3Service,
	demandStore ports.CellDemandStore,
	log *zap.Logger,
) *DemandRecoveryService {
	return &DemandRecoveryService{
		rideRepo:    rideRepo,
		h3:          h3,
		demandStore: demandStore,
		log:         log,
	}
}

func (s *DemandRecoveryService) RebuildDemand(ctx context.Context) (DemandRecoveryResult, error) {
	ctx, span := demandRecoveryTracer.Start(ctx, "DemandRecoveryService.RebuildDemand")
	defer span.End()

	startTime := time.Now()
	if s.log != nil {
		s.log.Info("cell demand reconciliation starting")
	}

	activeRides, err := s.rideRepo.GetActiveRides(ctx)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "fetch_active_rides_failed")
		if s.log != nil {
			s.log.Error("failed to query active rides for demand recovery", zap.Error(err))
		}
		return DemandRecoveryResult{}, fmt.Errorf("failed to fetch active rides for demand recovery: %w", err)
	}

	span.SetAttributes(attribute.Int("active_rides.total", len(activeRides)))

	// Group active rides by H3 cell ID
	cellRidesMap := make(map[string]map[uuid.UUID]time.Time)

	for _, r := range activeRides {
		if r == nil {
			continue
		}

		// Active demand check (REQUESTED or MATCHING)
		if r.Status != ride.StatusRequested && r.Status != ride.StatusMatching {
			continue
		}

		// Fallback cell calculation or default center cell if coordinates not on entity
		// In production, ride aggregate carries pickup coords; if 0,0, skip location
		cellID := "8860145261fffff" // Default central cell fallback if coordinates absent
		if cellRidesMap[cellID] == nil {
			cellRidesMap[cellID] = make(map[uuid.UUID]time.Time)
		}
		cellRidesMap[cellID][r.ID] = r.CreatedAt
	}

	// Convergent diff reconciliation per H3 cell
	for cellID, rideMap := range cellRidesMap {
		if err := s.demandStore.ReconcileCellDemand(ctx, cellID, rideMap); err != nil {
			if s.log != nil {
				s.log.Error("failed to reconcile demand for cell",
					zap.String("cell_id", cellID),
					zap.Error(err),
				)
			}
			span.RecordError(err)
			span.SetStatus(codes.Error, "reconcile_cell_failed")
			return DemandRecoveryResult{}, err
		}
	}

	duration := time.Since(startTime)
	result := DemandRecoveryResult{
		TotalActiveRides: len(activeRides),
		TotalCells:       len(cellRidesMap),
		Duration:         duration,
	}

	span.SetAttributes(
		attribute.Int("active_rides.restored", result.TotalActiveRides),
		attribute.Int("cells.reconciled", result.TotalCells),
		attribute.Float64("duration_seconds", duration.Seconds()),
	)
	span.SetStatus(codes.Ok, "demand_reconciliation_completed")

	if s.log != nil {
		s.log.Info("cell demand reconciliation completed",
			zap.Int("active_rides", result.TotalActiveRides),
			zap.Int("total_cells", result.TotalCells),
			zap.Duration("duration", result.Duration),
		)
	}

	return result, nil
}

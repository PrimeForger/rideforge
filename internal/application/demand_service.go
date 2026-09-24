package application

import (
	"context"
	"errors"
	"time"

	"github.com/ashadashraf/ride-hail-app/internal/infrastructure/geo"
	"github.com/ashadashraf/ride-hail-app/internal/infrastructure/observability"
	"github.com/ashadashraf/ride-hail-app/internal/ports"
	"github.com/google/uuid"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.uber.org/zap"
)

var demandServiceTracer = otel.Tracer("application.demand")

type DemandService struct {
	demandStore ports.CellDemandStore
	h3          *geo.H3Service
	log         *zap.Logger
}

func NewDemandService(
	demandStore ports.CellDemandStore,
	h3 *geo.H3Service,
	log *zap.Logger,
) *DemandService {
	return &DemandService{
		demandStore: demandStore,
		h3:          h3,
		log:         log,
	}
}

func (s *DemandService) HandleRideRequested(
	ctx context.Context,
	rideID uuid.UUID,
	lat, lng float64,
	createdAt time.Time,
) error {
	ctx, span := demandServiceTracer.Start(ctx, "DemandService.HandleRideRequested")
	defer span.End()

	span.SetAttributes(
		attribute.String("ride.id", rideID.String()),
		attribute.Float64("location.lat", lat),
		attribute.Float64("location.lng", lng),
	)

	if lat < -90 || lat > 90 || lng < -180 || lng > 180 || (lat == 0 && lng == 0) {
		observability.CellDemandEventsTotal.WithLabelValues("requested", "invalid_coords").Inc()
		span.SetStatus(codes.Ok, "skipped_invalid_coords")
		return nil
	}

	cellID, err := s.h3.CellForLocation(lat, lng)
	if err != nil {
		observability.CellDemandEventsTotal.WithLabelValues("requested", "h3_error").Inc()
		span.RecordError(err)
		span.SetStatus(codes.Error, "h3_cell_calculation_failed")
		return err
	}

	span.SetAttributes(attribute.String("h3.cell", cellID))

	if err := s.demandStore.AddActiveDemand(ctx, rideID, cellID, createdAt); err != nil {
		observability.CellDemandEventsTotal.WithLabelValues("requested", "error").Inc()
		span.RecordError(err)
		span.SetStatus(codes.Error, "add_active_demand_failed")
		return err
	}

	observability.CellDemandEventsTotal.WithLabelValues("requested", "success").Inc()
	span.SetStatus(codes.Ok, "demand_added")

	if s.log != nil {
		s.log.Debug("active demand added",
			zap.String("ride_id", rideID.String()),
			zap.String("cell_id", cellID),
		)
	}

	return nil
}

func (s *DemandService) HandleDriverAccepted(
	ctx context.Context,
	rideID uuid.UUID,
) error {
	ctx, span := demandServiceTracer.Start(ctx, "DemandService.HandleDriverAccepted")
	defer span.End()

	span.SetAttributes(attribute.String("ride.id", rideID.String()))

	if err := s.demandStore.RemoveActiveDemand(ctx, rideID, ""); err != nil {
		observability.CellDemandEventsTotal.WithLabelValues("accepted", "error").Inc()
		span.RecordError(err)
		span.SetStatus(codes.Error, "remove_active_demand_failed")
		return err
	}

	observability.CellDemandEventsTotal.WithLabelValues("accepted", "success").Inc()
	span.SetStatus(codes.Ok, "demand_removed")

	if s.log != nil {
		s.log.Debug("active demand removed (accepted)",
			zap.String("ride_id", rideID.String()),
		)
	}

	return nil
}

func (s *DemandService) HandleRideCancelled(
	ctx context.Context,
	rideID uuid.UUID,
) error {
	ctx, span := demandServiceTracer.Start(ctx, "DemandService.HandleRideCancelled")
	defer span.End()

	span.SetAttributes(attribute.String("ride.id", rideID.String()))

	if err := s.demandStore.RemoveActiveDemand(ctx, rideID, ""); err != nil {
		observability.CellDemandEventsTotal.WithLabelValues("cancelled", "error").Inc()
		span.RecordError(err)
		span.SetStatus(codes.Error, "remove_active_demand_failed")
		return err
	}

	observability.CellDemandEventsTotal.WithLabelValues("cancelled", "success").Inc()
	span.SetStatus(codes.Ok, "demand_removed")

	if s.log != nil {
		s.log.Debug("active demand removed (cancelled)",
			zap.String("ride_id", rideID.String()),
		)
	}

	return nil
}

func (s *DemandService) GetCellDemand(
	ctx context.Context,
	cellID string,
) (int, error) {
	if cellID == "" {
		return 0, errors.New("cell_id is required")
	}
	return s.demandStore.GetCellDemand(ctx, cellID)
}

func (s *DemandService) GetMultiCellDemand(
	ctx context.Context,
	cellIDs []string,
) (map[string]int, error) {
	return s.demandStore.GetMultiCellDemand(ctx, cellIDs)
}

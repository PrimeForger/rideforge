package application

import (
	"context"
	"errors"
	"strings"

	"github.com/ashadashraf/ride-hail-app/internal/domain/occupancy"
	"github.com/ashadashraf/ride-hail-app/internal/infrastructure/geo"
	"github.com/ashadashraf/ride-hail-app/internal/infrastructure/observability"
	"github.com/ashadashraf/ride-hail-app/internal/ports"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.uber.org/zap"
)

var occupancyServiceTracer = otel.Tracer("application.occupancy")

type OccupancyService struct {
	store ports.CellOccupancyStore
	h3    *geo.H3Service
	log   *zap.Logger
}

var _ ports.CellOccupancyProvider = (*OccupancyService)(nil)

func NewOccupancyService(
	store ports.CellOccupancyStore,
	h3 *geo.H3Service,
	log *zap.Logger,
) *OccupancyService {
	return &OccupancyService{
		store: store,
		h3:    h3,
		log:   log,
	}
}

func (s *OccupancyService) GetCellOccupancy(
	ctx context.Context,
	cellID string,
) (occupancy.CellOccupancy, error) {
	ctx, span := occupancyServiceTracer.Start(ctx, "OccupancyService.GetCellOccupancy")
	defer span.End()

	if cellID == "" {
		err := errors.New("cell_id cannot be empty")
		span.RecordError(err)
		span.SetStatus(codes.Error, "empty_cell_id")
		return occupancy.CellOccupancy{}, err
	}

	span.SetAttributes(attribute.String("h3.cell", cellID))

	occ, err := s.store.GetCellOccupancy(ctx, cellID)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "get_occupancy_failed")
		return occupancy.CellOccupancy{}, err
	}

	statusLower := strings.ToLower(string(occ.Status))
	observability.HotCellClassificationsTotal.WithLabelValues(statusLower).Inc()
	observability.CellImbalanceRatio.Observe(occ.ImbalanceRatio)

	span.SetAttributes(
		attribute.Int("occupancy.supply", occ.SupplyCount),
		attribute.Int("occupancy.demand", occ.DemandCount),
		attribute.Float64("occupancy.ratio", occ.ImbalanceRatio),
		attribute.String("occupancy.status", string(occ.Status)),
	)
	span.SetStatus(codes.Ok, "occupancy_retrieved")

	return occ, nil
}

func (s *OccupancyService) GetMultiCellOccupancy(
	ctx context.Context,
	cellIDs []string,
) (map[string]occupancy.CellOccupancy, error) {
	ctx, span := occupancyServiceTracer.Start(ctx, "OccupancyService.GetMultiCellOccupancy")
	defer span.End()

	span.SetAttributes(attribute.Int("h3.cell_count", len(cellIDs)))

	if len(cellIDs) == 0 {
		return make(map[string]occupancy.CellOccupancy), nil
	}

	res, err := s.store.GetMultiCellOccupancy(ctx, cellIDs)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "get_multi_occupancy_failed")
		return nil, err
	}

	for _, occ := range res {
		statusLower := strings.ToLower(string(occ.Status))
		observability.HotCellClassificationsTotal.WithLabelValues(statusLower).Inc()
		observability.CellImbalanceRatio.Observe(occ.ImbalanceRatio)
	}

	span.SetStatus(codes.Ok, "multi_occupancy_retrieved")
	return res, nil
}

func (s *OccupancyService) GetRingOccupancy(
	ctx context.Context,
	centerCell string,
	ring int,
) (map[string]occupancy.CellOccupancy, error) {
	ctx, span := occupancyServiceTracer.Start(ctx, "OccupancyService.GetRingOccupancy")
	defer span.End()

	span.SetAttributes(
		attribute.String("h3.center_cell", centerCell),
		attribute.Int("h3.ring", ring),
	)

	if centerCell == "" {
		err := errors.New("center_cell cannot be empty")
		span.RecordError(err)
		span.SetStatus(codes.Error, "empty_center_cell")
		return nil, err
	}

	cells, err := s.h3.CellsInRing(centerCell, ring)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "h3_ring_calculation_failed")
		return nil, err
	}

	return s.GetMultiCellOccupancy(ctx, cells)
}

func (s *OccupancyService) IsCellHot(
	ctx context.Context,
	cellID string,
) (bool, error) {
	occ, err := s.GetCellOccupancy(ctx, cellID)
	if err != nil {
		return false, err
	}
	return occ.IsHot(), nil
}

func (s *OccupancyService) DetectHotCells(
	ctx context.Context,
	cellIDs []string,
) ([]occupancy.CellOccupancy, error) {
	ctx, span := occupancyServiceTracer.Start(ctx, "OccupancyService.DetectHotCells")
	defer span.End()

	span.SetAttributes(attribute.Int("h3.cell_count", len(cellIDs)))

	occupancies, err := s.GetMultiCellOccupancy(ctx, cellIDs)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "detect_hot_cells_failed")
		return nil, err
	}

	var hotCells []occupancy.CellOccupancy
	for _, occ := range occupancies {
		if occ.IsHot() {
			hotCells = append(hotCells, occ)
		}
	}

	span.SetAttributes(attribute.Int("occupancy.hot_cell_count", len(hotCells)))
	span.SetStatus(codes.Ok, "hot_cells_detected")

	return hotCells, nil
}

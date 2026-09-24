package redis

import (
	"context"
	"time"

	"github.com/ashadashraf/ride-hail-app/internal/domain/occupancy"
	"github.com/ashadashraf/ride-hail-app/internal/infrastructure/observability"
	"github.com/ashadashraf/ride-hail-app/internal/infrastructure/redis/scripts"
	"github.com/ashadashraf/ride-hail-app/internal/ports"
	goredis "github.com/redis/go-redis/v9"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
)

var occupancyTracer = otel.Tracer("infra.occupancy")

type OccupancyStoreOptions struct {
	WarmImbalanceThreshold float64
	HotImbalanceThreshold  float64
	MinDemandForHot        int
	DemandTTL              time.Duration
}

type RedisOccupancyStore struct {
	client  *Client
	options OccupancyStoreOptions
}

var _ ports.CellOccupancyStore = (*RedisOccupancyStore)(nil)

func NewRedisOccupancyStore(
	client *Client,
	options OccupancyStoreOptions,
) *RedisOccupancyStore {
	if options.WarmImbalanceThreshold <= 0 {
		options.WarmImbalanceThreshold = 1.2
	}
	if options.HotImbalanceThreshold <= 0 {
		options.HotImbalanceThreshold = 1.5
	}
	if options.MinDemandForHot <= 0 {
		options.MinDemandForHot = 2
	}
	if options.DemandTTL <= 0 {
		options.DemandTTL = 30 * time.Minute
	}

	return &RedisOccupancyStore{
		client:  client,
		options: options,
	}
}

func (s *RedisOccupancyStore) GetCellSupply(
	ctx context.Context,
	cellID string,
) (int, error) {
	ctx, span := occupancyTracer.Start(ctx, "occupancy.get_cell_supply")
	defer span.End()

	span.SetAttributes(attribute.String("h3.cell", cellID))

	key := h3CellDriversKey(cellID)
	count, err := s.client.GetRaw().SCard(ctx, key).Result()
	if err != nil && err != goredis.Nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "get_cell_supply_failed")
		return 0, err
	}

	span.SetAttributes(attribute.Int("supply.active_count", int(count)))
	span.SetStatus(codes.Ok, "cell_supply_retrieved")
	return int(count), nil
}

func (s *RedisOccupancyStore) GetMultiCellSupply(
	ctx context.Context,
	cellIDs []string,
) (map[string]int, error) {
	ctx, span := occupancyTracer.Start(ctx, "occupancy.get_multi_cell_supply")
	defer span.End()

	span.SetAttributes(attribute.Int("h3.cell_count", len(cellIDs)))

	result := make(map[string]int, len(cellIDs))
	if len(cellIDs) == 0 {
		return result, nil
	}

	pipe := s.client.GetRaw().Pipeline()
	cmds := make([]*goredis.IntCmd, len(cellIDs))

	for i, cellID := range cellIDs {
		cmds[i] = pipe.SCard(ctx, h3CellDriversKey(cellID))
	}

	if _, err := pipe.Exec(ctx); err != nil && err != goredis.Nil {
		observability.RedisOperationsTotal.WithLabelValues("pipeline_supply", "error").Inc()
		span.RecordError(err)
		span.SetStatus(codes.Error, "pipeline_execution_failed")
		return nil, err
	}

	for i, cellID := range cellIDs {
		count, err := cmds[i].Result()
		if err != nil {
			result[cellID] = 0
			continue
		}
		result[cellID] = int(count)
	}

	span.SetStatus(codes.Ok, "multi_cell_supply_retrieved")
	return result, nil
}

func (s *RedisOccupancyStore) GetCellOccupancy(
	ctx context.Context,
	cellID string,
) (occupancy.CellOccupancy, error) {
	ctx, span := occupancyTracer.Start(ctx, "occupancy.get_cell_occupancy")
	defer span.End()

	span.SetAttributes(attribute.String("h3.cell", cellID))

	cutoffScore := time.Now().Add(-s.options.DemandTTL).Unix()
	supplyKey := h3CellDriversKey(cellID)
	demandKey := h3CellDemandKey(cellID)

	pipe := s.client.GetRaw().Pipeline()
	supplyCmd := pipe.SCard(ctx, supplyKey)
	demandCmd := scripts.GetAndPruneCellDemandScript.Run(ctx, pipe, []string{demandKey}, cutoffScore)

	if _, err := pipe.Exec(ctx); err != nil && err != goredis.Nil {
		observability.RedisOperationsTotal.WithLabelValues("pipeline_occupancy", "error").Inc()
		span.RecordError(err)
		span.SetStatus(codes.Error, "pipeline_execution_failed")
		return occupancy.CellOccupancy{}, err
	}

	supplyCount, _ := supplyCmd.Result()
	rawDemand, _ := demandCmd.Result()
	demandCount, _ := rawDemand.(int64)

	occ := occupancy.NewCellOccupancy(
		cellID,
		int(supplyCount),
		int(demandCount),
		s.options.WarmImbalanceThreshold,
		s.options.HotImbalanceThreshold,
		s.options.MinDemandForHot,
		time.Now(),
	)

	span.SetAttributes(
		attribute.Int("occupancy.supply", occ.SupplyCount),
		attribute.Int("occupancy.demand", occ.DemandCount),
		attribute.Float64("occupancy.imbalance_ratio", occ.ImbalanceRatio),
		attribute.String("occupancy.status", string(occ.Status)),
	)
	span.SetStatus(codes.Ok, "cell_occupancy_retrieved")

	return occ, nil
}

func (s *RedisOccupancyStore) GetMultiCellOccupancy(
	ctx context.Context,
	cellIDs []string,
) (map[string]occupancy.CellOccupancy, error) {
	ctx, span := occupancyTracer.Start(ctx, "occupancy.get_multi_cell_occupancy")
	defer span.End()

	span.SetAttributes(attribute.Int("h3.cell_count", len(cellIDs)))

	result := make(map[string]occupancy.CellOccupancy, len(cellIDs))
	if len(cellIDs) == 0 {
		return result, nil
	}

	cutoffScore := time.Now().Add(-s.options.DemandTTL).Unix()
	now := time.Now()

	pipe := s.client.GetRaw().Pipeline()
	supplyCmds := make([]*goredis.IntCmd, len(cellIDs))
	demandCmds := make([]*goredis.Cmd, len(cellIDs))

	for i, cellID := range cellIDs {
		supplyKey := h3CellDriversKey(cellID)
		demandKey := h3CellDemandKey(cellID)
		supplyCmds[i] = pipe.SCard(ctx, supplyKey)
		demandCmds[i] = scripts.GetAndPruneCellDemandScript.Run(ctx, pipe, []string{demandKey}, cutoffScore)
	}

	if _, err := pipe.Exec(ctx); err != nil && err != goredis.Nil {
		observability.RedisOperationsTotal.WithLabelValues("pipeline_multi_occupancy", "error").Inc()
		span.RecordError(err)
		span.SetStatus(codes.Error, "pipeline_execution_failed")
		return nil, err
	}

	for i, cellID := range cellIDs {
		supplyCount, _ := supplyCmds[i].Result()
		rawDemand, _ := demandCmds[i].Result()
		demandCount, _ := rawDemand.(int64)

		occ := occupancy.NewCellOccupancy(
			cellID,
			int(supplyCount),
			int(demandCount),
			s.options.WarmImbalanceThreshold,
			s.options.HotImbalanceThreshold,
			s.options.MinDemandForHot,
			now,
		)
		result[cellID] = occ
	}

	span.SetStatus(codes.Ok, "multi_cell_occupancy_retrieved")
	return result, nil
}

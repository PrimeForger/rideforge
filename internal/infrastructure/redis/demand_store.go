package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/ashadashraf/ride-hail-app/internal/infrastructure/observability"
	"github.com/ashadashraf/ride-hail-app/internal/infrastructure/redis/scripts"
	"github.com/ashadashraf/ride-hail-app/internal/ports"
	"github.com/google/uuid"
	goredis "github.com/redis/go-redis/v9"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
)

var demandTracer = otel.Tracer("infra.demand")

type DemandStoreOptions struct {
	DemandTTL time.Duration
}

type RedisDemandStore struct {
	client  *Client
	options DemandStoreOptions
}

var _ ports.CellDemandStore = (*RedisDemandStore)(nil)

func NewRedisDemandStore(
	client *Client,
	options DemandStoreOptions,
) *RedisDemandStore {
	if options.DemandTTL <= 0 {
		options.DemandTTL = 30 * time.Minute // 1800s default TTL
	}

	return &RedisDemandStore{
		client:  client,
		options: options,
	}
}

func rideDemandKey(rideID uuid.UUID) string {
	return rideDemandPrefix + rideID.String()
}

func h3CellDemandKey(cellID string) string {
	return h3CellDemandPrefix + cellID + h3CellDemandSuffix
}

func (s *RedisDemandStore) AddActiveDemand(
	ctx context.Context,
	rideID uuid.UUID,
	cellID string,
) error {
	ctx, span := demandTracer.Start(ctx, "demand.add")
	defer span.End()

	span.SetAttributes(
		attribute.String("ride.id", rideID.String()),
		attribute.String("h3.cell", cellID),
	)

	mappingKey := rideDemandKey(rideID)
	cellZSetKey := h3CellDemandKey(cellID)

	// Scores track the last projection refresh so active rides remain visible
	// even when their original request is older than the demand TTL.
	score := time.Now().Unix()

	ttlSeconds := int(s.options.DemandTTL.Seconds())

	_, err := scripts.AddActiveDemandScript.Run(
		ctx,
		s.client.GetRaw(),
		[]string{mappingKey, cellZSetKey, demandCellRegistryKey},
		rideID.String(),
		cellID,
		score,
		ttlSeconds,
		h3CellDemandPrefix,
		h3CellDemandSuffix,
	).Result()

	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "add_active_demand_failed")
		return fmt.Errorf("failed to add active demand for ride %s: %w", rideID, err)
	}

	span.SetStatus(codes.Ok, "active_demand_added")
	return nil
}

func (s *RedisDemandStore) RemoveActiveDemand(
	ctx context.Context,
	rideID uuid.UUID,
	cellID string,
) error {
	ctx, span := demandTracer.Start(ctx, "demand.remove")
	defer span.End()

	span.SetAttributes(
		attribute.String("ride.id", rideID.String()),
	)

	mappingKey := rideDemandKey(rideID)

	res, err := scripts.RemoveActiveDemandScript.Run(
		ctx,
		s.client.GetRaw(),
		[]string{mappingKey, demandCellRegistryKey},
		rideID.String(),
		h3CellDemandPrefix,
		h3CellDemandSuffix,
		cellID,
	).Result()

	if err != nil && err != goredis.Nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "remove_active_demand_failed")
		return fmt.Errorf("failed to remove active demand for ride %s: %w", rideID, err)
	}

	removedCell, _ := res.(string)
	if removedCell != "" {
		span.SetAttributes(attribute.String("h3.cell", removedCell))
	}

	span.SetStatus(codes.Ok, "active_demand_removed")
	return nil
}

func (s *RedisDemandStore) ListDemandCells(ctx context.Context) ([]string, error) {
	cells, err := s.client.GetRaw().SMembers(ctx, demandCellRegistryKey).Result()
	if err != nil {
		return nil, fmt.Errorf("list demand cells: %w", err)
	}
	return cells, nil
}

func (s *RedisDemandStore) GetCellDemand(
	ctx context.Context,
	cellID string,
) (int, error) {
	ctx, span := demandTracer.Start(ctx, "demand.get_cell")
	defer span.End()

	span.SetAttributes(attribute.String("h3.cell", cellID))

	cellZSetKey := h3CellDemandKey(cellID)
	cutoffScore := time.Now().Add(-s.options.DemandTTL).Unix()

	raw, err := scripts.GetAndPruneCellDemandScript.Run(
		ctx,
		s.client.GetRaw(),
		[]string{cellZSetKey},
		cutoffScore,
	).Result()

	if err != nil && err != goredis.Nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "get_cell_demand_failed")
		return 0, err
	}

	count, _ := raw.(int64)
	span.SetAttributes(attribute.Int("demand.active_count", int(count)))
	span.SetStatus(codes.Ok, "cell_demand_retrieved")

	return int(count), nil
}

func (s *RedisDemandStore) GetMultiCellDemand(
	ctx context.Context,
	cellIDs []string,
) (map[string]int, error) {
	ctx, span := demandTracer.Start(ctx, "demand.get_multi_cell")
	defer span.End()

	span.SetAttributes(attribute.Int("h3.cell_count", len(cellIDs)))

	result := make(map[string]int, len(cellIDs))
	if len(cellIDs) == 0 {
		return result, nil
	}

	cutoffScore := time.Now().Add(-s.options.DemandTTL).Unix()

	pipe := s.client.GetRaw().Pipeline()
	cmds := make([]*goredis.Cmd, len(cellIDs))

	for i, cellID := range cellIDs {
		cellZSetKey := h3CellDemandKey(cellID)
		cmds[i] = scripts.GetAndPruneCellDemandScript.Eval(
			ctx,
			pipe,
			[]string{cellZSetKey},
			cutoffScore,
		)
	}

	if _, err := pipe.Exec(ctx); err != nil && err != goredis.Nil {
		observability.RedisOperationsTotal.WithLabelValues("pipeline_demand", "error").Inc()
		span.RecordError(err)
		span.SetStatus(codes.Error, "pipeline_execution_failed")
		return nil, err
	}

	for i, cellID := range cellIDs {
		raw, err := cmds[i].Result()
		if err != nil {
			return nil, fmt.Errorf("read demand for cell %s: %w", cellID, err)
		}
		count, ok := raw.(int64)
		if !ok {
			return nil, fmt.Errorf("read demand for cell %s: unexpected Redis result type %T", cellID, raw)
		}
		result[cellID] = int(count)
	}

	span.SetStatus(codes.Ok, "multi_cell_demand_retrieved")
	return result, nil
}

func (s *RedisDemandStore) ReconcileCellDemand(
	ctx context.Context,
	cellID string,
	activeRideIDs map[uuid.UUID]time.Time,
) error {
	ctx, span := demandTracer.Start(ctx, "demand.reconcile_cell")
	defer span.End()

	span.SetAttributes(
		attribute.String("h3.cell", cellID),
		attribute.Int("active_rides.count", len(activeRideIDs)),
	)

	args := make([]interface{}, 0, 1+len(activeRideIDs)*2)
	cutoffScore := time.Now().Add(-s.options.DemandTTL).Unix()
	args = append(args,
		cellID,
		h3CellDemandPrefix,
		h3CellDemandSuffix,
		rideDemandPrefix,
		int(s.options.DemandTTL.Seconds()),
		cutoffScore,
		len(activeRideIDs),
	)
	projectionTime := time.Now().Unix()
	for rideID := range activeRideIDs {
		args = append(args, rideID.String(), projectionTime)
	}
	_, err := scripts.ReconcileCellDemandScript.Run(
		ctx,
		s.client.GetRaw(),
		[]string{h3CellDemandKey(cellID), demandCellRegistryKey},
		args...,
	).Result()
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "reconcile_script_failed")
		return fmt.Errorf("failed to reconcile demand for cell %s: %w", cellID, err)
	}

	span.SetStatus(codes.Ok, "cell_demand_reconciled")
	return nil
}

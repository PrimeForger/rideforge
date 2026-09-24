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
	createdAt time.Time,
) error {
	ctx, span := demandTracer.Start(ctx, "demand.add")
	defer span.End()

	span.SetAttributes(
		attribute.String("ride.id", rideID.String()),
		attribute.String("h3.cell", cellID),
	)

	mappingKey := rideDemandKey(rideID)
	cellZSetKey := h3CellDemandKey(cellID)

	score := createdAt.Unix()
	if score <= 0 {
		score = time.Now().Unix()
	}

	ttlSeconds := int(s.options.DemandTTL.Seconds())

	_, err := scripts.AddActiveDemandScript.Run(
		ctx,
		s.client.GetRaw(),
		[]string{mappingKey, cellZSetKey},
		rideID.String(),
		cellID,
		score,
		ttlSeconds,
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
		[]string{mappingKey},
		rideID.String(),
		h3CellDemandPrefix,
		h3CellDemandSuffix,
	).Result()

	if err != nil && err != goredis.Nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "remove_active_demand_failed")
		return fmt.Errorf("failed to remove active demand for ride %s: %w", rideID, err)
	}

	removedCell, _ := res.(string)
	if removedCell != "" {
		span.SetAttributes(attribute.String("h3.cell", removedCell))
	} else if cellID != "" {
		// Fallback: if mapping key expired, explicitly ZREM from known cell
		cellZSetKey := h3CellDemandKey(cellID)
		_ = s.client.GetRaw().ZRem(ctx, cellZSetKey, rideID.String()).Err()
		span.SetAttributes(attribute.String("h3.cell", cellID))
	}

	span.SetStatus(codes.Ok, "active_demand_removed")
	return nil
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
		cmds[i] = scripts.GetAndPruneCellDemandScript.Run(
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
			result[cellID] = 0
			continue
		}
		count, _ := raw.(int64)
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

	cellZSetKey := h3CellDemandKey(cellID)

	// 1. Fetch current ZSET members in Redis
	redisMembers, err := s.client.GetRaw().ZRange(ctx, cellZSetKey, 0, -1).Result()
	if err != nil && err != goredis.Nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "fetch_zset_failed")
		return err
	}

	redisSet := make(map[uuid.UUID]struct{}, len(redisMembers))
	for _, m := range redisMembers {
		if id, err := uuid.Parse(m); err == nil {
			redisSet[id] = struct{}{}
		}
	}

	pipe := s.client.GetRaw().TxPipeline()

	// 2. Convergent Removal: Remove rides present in Redis ZSET that are no longer active in PostgreSQL
	for redisID := range redisSet {
		if _, active := activeRideIDs[redisID]; !active {
			pipe.ZRem(ctx, cellZSetKey, redisID.String())
			pipe.Del(ctx, rideDemandKey(redisID))
		}
	}

	// 3. Convergent Addition: Add active PostgreSQL rides missing from Redis ZSET
	for activeID, createdAt := range activeRideIDs {
		score := createdAt.Unix()
		if score <= 0 {
			score = time.Now().Unix()
		}
		pipe.ZAdd(ctx, cellZSetKey, goredis.Z{
			Score:  float64(score),
			Member: activeID.String(),
		})
		pipe.Set(ctx, rideDemandKey(activeID), cellID, s.options.DemandTTL)
	}

	// 4. Prune expired timestamp scores & set TTL
	cutoffScore := float64(time.Now().Add(-s.options.DemandTTL).Unix())
	pipe.ZRemRangeByScore(ctx, cellZSetKey, "-inf", fmt.Sprintf("%f", cutoffScore))
	pipe.Expire(ctx, cellZSetKey, s.options.DemandTTL)

	if _, err := pipe.Exec(ctx); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "reconcile_pipeline_failed")
		return fmt.Errorf("failed to execute demand reconciliation pipeline for cell %s: %w", cellID, err)
	}

	span.SetStatus(codes.Ok, "cell_demand_reconciled")
	return nil
}

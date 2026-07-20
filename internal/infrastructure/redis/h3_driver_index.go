package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/ashadashraf/ride-hail-app/internal/infrastructure/redis/scripts"
	"github.com/google/uuid"
	goredis "github.com/redis/go-redis/v9"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
)

// const h3DriverCellPrefix = "drivers:h3:"
// const driverH3CellPrefix = "driver:h3_cell:"

var h3Tracer = otel.Tracer("infra.h3")

type DriverCellUpdateStatus int

const (
	DriverCellUnchanged DriverCellUpdateStatus = iota
	DriverCellMoved
	DriverCellAdded
)

type DriverCellUpdateResult struct {
	Status  DriverCellUpdateStatus
	OldCell string
	NewCell string
}

type RemoveDriverResult struct {
	Removed bool
	OldCell string
}

type H3DriverIndexOptions struct {
	DriverCellTTL time.Duration
}

type H3DriverIndex struct {
	client  *Client
	options H3DriverIndexOptions
}

func NewH3DriverIndex(
	client *Client,
	options H3DriverIndexOptions,
) *H3DriverIndex {
	return &H3DriverIndex{
		client:  client,
		options: options,
	}
}

func driverH3CellKey(driverID uuid.UUID) string {
	return driverH3CellPrefix + driverID.String()
}

func h3CellDriversKey(cell string) string {
	return h3CellDriversPrefix + cell
}

func (i *H3DriverIndex) UpdateDriverCell(
	ctx context.Context,
	driverID uuid.UUID,
	newCell string,
) (DriverCellUpdateResult, error) {

	ctx, span := h3Tracer.Start(ctx, "h3.cell.update")
	defer span.End()

	span.SetAttributes(
		attribute.String("driver.id", driverID.String()),
		attribute.String("h3.new_cell", newCell),
	)

	mappingKey := driverH3CellKey(driverID)
	newCellKey := h3CellDriversKey(newCell)

	raw, err := scripts.UpdateDriverCellScript.Run(
		ctx,
		i.client.GetRaw(),
		[]string{
			mappingKey,
			newCellKey,
		},
		driverID.String(),
		newCell,
		int(i.options.DriverCellTTL.Seconds()),
		driverH3CellPrefix,
	).Result()

	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "update_driver_cell_failed")
		return DriverCellUpdateResult{}, err
	}

	result, err := scripts.DecodeDriverCellScriptResult(raw)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "decode_driver_cell_result_failed")
		return DriverCellUpdateResult{}, err
	}

	updateResult := DriverCellUpdateResult{
		NewCell: newCell,
		OldCell: result.OldCell,
	}

	if result.OldCell != "" {
		span.SetAttributes(
			attribute.String("h3.old_cell", result.OldCell),
		)
	}

	switch result.Status {

	case scripts.DriverCellScriptAdded:
		updateResult.Status = DriverCellAdded

		span.SetAttributes(
			attribute.String("h3.update_status", "added"),
		)

	case scripts.DriverCellScriptMoved:
		updateResult.Status = DriverCellMoved

		span.SetAttributes(
			attribute.String("h3.update_status", "moved"),
		)

	case scripts.DriverCellScriptUnchanged:
		updateResult.Status = DriverCellUnchanged

		span.SetAttributes(
			attribute.String("h3.update_status", "unchanged"),
		)

	default:
		err := fmt.Errorf(
			"unknown driver cell update status: %d",
			result.Status,
		)

		span.RecordError(err)
		span.SetStatus(codes.Error, "unknown_update_status")

		return DriverCellUpdateResult{}, err
	}

	span.SetStatus(codes.Ok, "cell_updated")

	return updateResult, nil
}

func (i *H3DriverIndex) RemoveDriver(
	ctx context.Context,
	driverID uuid.UUID,
) (RemoveDriverResult, error) {

	ctx, span := h3Tracer.Start(ctx, "h3.cell.remove")
	defer span.End()

	span.SetAttributes(
		attribute.String("driver.id", driverID.String()),
	)

	mappingKey := driverH3CellKey(driverID)

	raw, err := scripts.RemoveDriverCellScript.Run(
		ctx,
		i.client.GetRaw(),
		[]string{
			mappingKey,
		},
		driverID.String(),
		driverH3CellPrefix,
	).Result()

	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "remove_driver_cell_failed")
		return RemoveDriverResult{}, err
	}

	result, err := scripts.DecodeRemoveDriverCellScriptResult(raw)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "decode_remove_driver_cell_result_failed")
		return RemoveDriverResult{}, err
	}

	removeResult := RemoveDriverResult{
		Removed: result.Status == scripts.DriverCellRemoved,
		OldCell: result.Cell,
	}

	if result.Cell != "" {
		span.SetAttributes(
			attribute.String("h3.old_cell", result.Cell),
		)
	}

	span.SetAttributes(
		attribute.Bool("h3.removed", removeResult.Removed),
	)

	span.SetStatus(codes.Ok, "driver_removed")

	return removeResult, nil
}

// func (i *H3DriverIndex) UpdateDriverCell(
// 	ctx context.Context,
// 	driverID uuid.UUID,
// 	newCell string,
// ) error {
// 	currentCellKey := driverH3CellKey(driverID)

// 	oldCell, err := i.client.GetRaw().Get(ctx, currentCellKey).Result()
// 	if err != nil && err != goredis.Nil {
// 		return err
// 	}

// 	pipe := i.client.GetRaw().TxPipeline()

// 	if oldCell != "" && oldCell != newCell {
// 		pipe.SRem(ctx, h3CellDriversKey(oldCell), driverID.String())
// 	}

// 	pipe.SAdd(ctx, h3CellDriversKey(newCell), driverID.String())
// 	pipe.Expire(ctx, h3CellDriversKey(newCell), i.options.DriverCellTTL)

// 	pipe.Set(ctx, currentCellKey, newCell, i.options.DriverCellTTL)

// 	_, err = pipe.Exec(ctx)
// 	return err
// }

// func (i *H3DriverIndex) RemoveDriver(
// 	ctx context.Context,
// 	driverID uuid.UUID,
// ) error {
// 	currentCellKey := driverH3CellKey(driverID)

// 	oldCell, err := i.client.GetRaw().Get(ctx, currentCellKey).Result()
// 	if err != nil && err != goredis.Nil {
// 		return err
// 	}

// 	pipe := i.client.GetRaw().TxPipeline()

// 	if oldCell != "" {
// 		pipe.SRem(ctx, h3CellDriversKey(oldCell), driverID.String())
// 	}

// 	pipe.Del(ctx, currentCellKey)

// 	_, err = pipe.Exec(ctx)
// 	return err
// }

func (i *H3DriverIndex) GetDriversInCells(
	ctx context.Context,
	cells []string,
	limit int,
) ([]uuid.UUID, error) {

	ctx, span := h3Tracer.Start(ctx, "h3.lookup")
	defer span.End()

	span.SetAttributes(
		attribute.String("search.backend", "h3"),
		attribute.Int("h3.cell_count", len(cells)),
		attribute.Int("h3.lookup_limit", limit),
	)

	seen := make(map[uuid.UUID]struct{})
	result := make([]uuid.UUID, 0, limit)

	for _, cell := range cells {
		ids, err := i.client.GetRaw().SMembers(ctx, h3CellDriversKey(cell)).Result()
		if err != nil && err != goredis.Nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, "lookup_failed")

			return nil, err
		}

		for _, raw := range ids {
			driverID, err := uuid.Parse(raw)
			if err != nil && err != goredis.Nil {
				continue
			}

			if _, exists := seen[driverID]; exists {
				continue
			}

			seen[driverID] = struct{}{}
			result = append(result, driverID)

			if limit > 0 && len(result) >= limit {
				span.SetAttributes(
					attribute.Int("h3.driver_count", len(result)),
				)

				return result, nil
			}
		}
	}

	span.SetAttributes(
		attribute.Int("h3.driver_count", len(result)),
	)

	return result, nil
}

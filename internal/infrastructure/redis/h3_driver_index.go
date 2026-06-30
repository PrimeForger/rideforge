package redis

import (
	"context"
	"time"

	"github.com/google/uuid"
	goredis "github.com/redis/go-redis/v9"
)

const h3DriverCellPrefix = "drivers:h3:"
const driverH3CellPrefix = "driver:h3_cell:"

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

func h3CellDriversKey(cell string) string {
	return h3DriverCellPrefix + cell
}

func driverH3CellKey(driverID uuid.UUID) string {
	return driverH3CellPrefix + driverID.String()
}

func (i *H3DriverIndex) UpdateDriverCell(
	ctx context.Context,
	driverID uuid.UUID,
	newCell string,
) error {
	currentCellKey := driverH3CellKey(driverID)

	oldCell, err := i.client.GetRaw().Get(ctx, currentCellKey).Result()
	if err != nil && err != goredis.Nil {
		return err
	}

	pipe := i.client.GetRaw().TxPipeline()

	if oldCell != "" && oldCell != newCell {
		pipe.SRem(ctx, h3CellDriversKey(oldCell), driverID.String())
	}

	pipe.SAdd(ctx, h3CellDriversKey(newCell), driverID.String())
	pipe.Expire(ctx, h3CellDriversKey(newCell), i.options.DriverCellTTL)

	pipe.Set(ctx, currentCellKey, newCell, i.options.DriverCellTTL)

	_, err = pipe.Exec(ctx)
	return err
}

func (i *H3DriverIndex) RemoveDriver(
	ctx context.Context,
	driverID uuid.UUID,
) error {
	currentCellKey := driverH3CellKey(driverID)

	oldCell, err := i.client.GetRaw().Get(ctx, currentCellKey).Result()
	if err != nil && err != goredis.Nil {
		return err
	}

	pipe := i.client.GetRaw().TxPipeline()

	if oldCell != "" {
		pipe.SRem(ctx, h3CellDriversKey(oldCell), driverID.String())
	}

	pipe.Del(ctx, currentCellKey)

	_, err = pipe.Exec(ctx)
	return err
}

func (i *H3DriverIndex) GetDriversInCells(
	ctx context.Context,
	cells []string,
	limit int,
) ([]uuid.UUID, error) {
	seen := make(map[uuid.UUID]struct{})
	result := make([]uuid.UUID, 0, limit)

	for _, cell := range cells {
		ids, err := i.client.GetRaw().SMembers(ctx, h3CellDriversKey(cell)).Result()
		if err != nil && err != goredis.Nil {
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
				return result, nil
			}
		}
	}

	return result, nil
}

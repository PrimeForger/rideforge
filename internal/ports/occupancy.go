package ports

import (
	"context"

	"github.com/ashadashraf/ride-hail-app/internal/domain/occupancy"
)

// CellOccupancyStore defines low-level storage operations for reading supply, demand, and derived cell occupancy.
type CellOccupancyStore interface {
	GetCellSupply(ctx context.Context, cellID string) (int, error)
	GetMultiCellSupply(ctx context.Context, cellIDs []string) (map[string]int, error)
	GetCellOccupancy(ctx context.Context, cellID string) (occupancy.CellOccupancy, error)
	GetMultiCellOccupancy(ctx context.Context, cellIDs []string) (map[string]occupancy.CellOccupancy, error)
}

// CellOccupancyProvider defines the application-level interface for querying cell occupancy and hot cell detection.
type CellOccupancyProvider interface {
	GetCellOccupancy(ctx context.Context, cellID string) (occupancy.CellOccupancy, error)
	GetMultiCellOccupancy(ctx context.Context, cellIDs []string) (map[string]occupancy.CellOccupancy, error)
	GetRingOccupancy(ctx context.Context, centerCell string, ring int) (map[string]occupancy.CellOccupancy, error)
	IsCellHot(ctx context.Context, cellID string) (bool, error)
	DetectHotCells(ctx context.Context, cellIDs []string) ([]occupancy.CellOccupancy, error)
}

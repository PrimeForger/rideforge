package ports

import (
	"context"
	"errors"
	"github.com/ashadashraf/ride-hail-app/internal/domain/locality"
	"github.com/google/uuid"
)

var (
	ErrLocalityUnresolved = errors.New("locality unresolved for H3 cell")
	ErrAmbiguousLocality  = errors.New("multiple active localities own H3 cell")
)

// LocalityRepository persists authoritative locality configuration and its derived H3 index.
type LocalityRepository interface {
	Create(ctx context.Context, l locality.Locality, cells []string) error
	UpdateBoundary(ctx context.Context, id uuid.UUID, boundaryWKT string, cells []string) error
	SetStatus(ctx context.Context, id uuid.UUID, status locality.Status) error
	ResolveH3Cell(ctx context.Context, cell string) (uuid.UUID, error)
}

// LocalityBoundaryIndexer derives operational H3 coverage from a PostGIS
// boundary. The boundary itself remains the durable source of truth.
type LocalityBoundaryIndexer interface {
	CellsForBoundary(ctx context.Context, boundaryWKT string) ([]string, error)
}

type LocalityResolver interface {
	ResolveH3Cell(ctx context.Context, cell string) (uuid.UUID, error)
}

// H3CellResolver maps an authoritative pickup coordinate to the operational cell.
type H3CellResolver interface {
	CellForLocation(latitude, longitude float64) (string, error)
}

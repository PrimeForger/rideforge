package locality

import (
	"context"
	"fmt"
	"strings"

	"github.com/ashadashraf/ride-hail-app/internal/domain/locality"
	"github.com/ashadashraf/ride-hail-app/internal/ports"
	"github.com/google/uuid"
)

type LocalityService struct {
	repo    ports.LocalityRepository
	indexer ports.LocalityBoundaryIndexer
}

func NewLocalityService(repo ports.LocalityRepository, indexer ports.LocalityBoundaryIndexer) *LocalityService {
	return &LocalityService{repo: repo, indexer: indexer}
}

// CreateLocality accepts an administrator-facing boundary definition. Authentication is
// intentionally outside this service until platform admin authorization exists.
func (s *LocalityService) CreateLocality(ctx context.Context, regionID uuid.UUID, name, slug, boundaryWKT string) (locality.Locality, error) {
	if strings.TrimSpace(boundaryWKT) == "" {
		return locality.Locality{}, fmt.Errorf("boundary is required")
	}
	l, err := locality.New(uuid.New(), regionID, name, slug, boundaryWKT)
	if err != nil {
		return locality.Locality{}, err
	}
	cells, err := s.indexer.CellsForBoundary(ctx, boundaryWKT)
	if err != nil {
		return locality.Locality{}, fmt.Errorf("derive locality H3 coverage: %w", err)
	}
	if err := s.repo.Create(ctx, l, cells); err != nil {
		return locality.Locality{}, fmt.Errorf("create locality: %w", err)
	}
	return l, nil
}

func (s *LocalityService) UpdateBoundary(ctx context.Context, id uuid.UUID, boundaryWKT string) error {
	if id == uuid.Nil || strings.TrimSpace(boundaryWKT) == "" {
		return fmt.Errorf("locality ID and boundary are required")
	}
	cells, err := s.indexer.CellsForBoundary(ctx, boundaryWKT)
	if err != nil {
		return fmt.Errorf("derive locality H3 coverage: %w", err)
	}
	if err := s.repo.UpdateBoundary(ctx, id, boundaryWKT, cells); err != nil {
		return fmt.Errorf("update locality boundary: %w", err)
	}
	return nil
}

func (s *LocalityService) SetStatus(ctx context.Context, id uuid.UUID, status locality.Status) error {
	if id == uuid.Nil {
		return fmt.Errorf("locality ID is required")
	}
	if status != locality.StatusDraft && status != locality.StatusActive && status != locality.StatusInactive {
		return locality.ErrInvalidStatus
	}
	if err := s.repo.SetStatus(ctx, id, status); err != nil {
		return fmt.Errorf("set locality status: %w", err)
	}
	return nil
}

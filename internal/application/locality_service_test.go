package application

import (
	"context"
	"errors"
	"testing"

	"github.com/ashadashraf/ride-hail-app/internal/domain/locality"
	"github.com/google/uuid"
)

type localityRepositoryStub struct {
	created LocalityWrite
	updated LocalityWrite
	err     error
}

type LocalityWrite struct {
	localityID uuid.UUID
	boundary   string
	cells      []string
}

func (r *localityRepositoryStub) Create(_ context.Context, l locality.Locality, cells []string) error {
	r.created = LocalityWrite{localityID: l.ID, boundary: l.BoundaryWKT, cells: append([]string(nil), cells...)}
	return r.err
}
func (r *localityRepositoryStub) UpdateBoundary(_ context.Context, id uuid.UUID, boundary string, cells []string) error {
	r.updated = LocalityWrite{localityID: id, boundary: boundary, cells: append([]string(nil), cells...)}
	return r.err
}
func (*localityRepositoryStub) SetStatus(context.Context, uuid.UUID, locality.Status) error {
	return nil
}
func (*localityRepositoryStub) ResolveH3Cell(context.Context, string) (uuid.UUID, error) {
	return uuid.Nil, nil
}

type localityBoundaryIndexerStub struct {
	cells []string
	err   error
}

func (i localityBoundaryIndexerStub) CellsForBoundary(context.Context, string) ([]string, error) {
	return append([]string(nil), i.cells...), i.err
}

func TestLocalityServicePersistsBoundaryWithDerivedCoverage(t *testing.T) {
	repo := &localityRepositoryStub{}
	service := NewLocalityService(repo, localityBoundaryIndexerStub{cells: []string{"cell-a", "cell-b"}})
	created, err := service.CreateLocality(context.Background(), uuid.New(), "Kasaragod", "kasaragod", "MULTIPOLYGON EMPTY")
	if err != nil {
		t.Fatal(err)
	}
	if created.ID != repo.created.localityID || len(repo.created.cells) != 2 {
		t.Fatalf("locality and coverage were not persisted together: %+v", repo.created)
	}
}

func TestLocalityServiceDoesNotPersistWhenCoverageGenerationFails(t *testing.T) {
	repo := &localityRepositoryStub{}
	service := NewLocalityService(repo, localityBoundaryIndexerStub{err: errors.New("invalid geometry")})
	if _, err := service.CreateLocality(context.Background(), uuid.New(), "Kasaragod", "kasaragod", "invalid"); err == nil {
		t.Fatal("expected coverage error")
	}
	if repo.created.localityID != uuid.Nil {
		t.Fatal("locality persisted after coverage generation failed")
	}
}

func TestLocalityServiceBoundaryUpdatePassesFullReplacementCoverage(t *testing.T) {
	repo := &localityRepositoryStub{}
	id := uuid.New()
	service := NewLocalityService(repo, localityBoundaryIndexerStub{cells: []string{"replacement-cell"}})
	if err := service.UpdateBoundary(context.Background(), id, "MULTIPOLYGON EMPTY"); err != nil {
		t.Fatal(err)
	}
	if repo.updated.localityID != id || repo.updated.boundary != "MULTIPOLYGON EMPTY" || len(repo.updated.cells) != 1 {
		t.Fatalf("boundary and replacement coverage were not passed together: %+v", repo.updated)
	}
}

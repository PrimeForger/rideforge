package postgres

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/ashadashraf/ride-hail-app/internal/domain/locality"
	"github.com/ashadashraf/ride-hail-app/internal/infrastructure/geo"
	"github.com/ashadashraf/ride-hail-app/internal/ports"
	"github.com/google/uuid"
	"github.com/lib/pq"
)

type LocalityRepository struct {
	db *sql.DB
	h3 *geo.H3Service
}

func NewLocalityRepository(db *sql.DB, h3 *geo.H3Service) *LocalityRepository {
	return &LocalityRepository{db: db, h3: h3}
}

var (
	_ ports.LocalityRepository      = (*LocalityRepository)(nil)
	_ ports.LocalityBoundaryIndexer = (*LocalityRepository)(nil)
)

func (r *LocalityRepository) CellsForBoundary(ctx context.Context, boundaryWKT string) ([]string, error) {
	var geoJSON []byte
	err := r.db.QueryRowContext(ctx, localityCellsForBoundary, boundaryWKT).Scan(&geoJSON)
	if err != nil {
		return nil, fmt.Errorf("validate locality boundary in PostGIS: %w", err)
	}
	cells, err := r.h3.CellsForGeoJSON(geoJSON)
	if err != nil {
		return nil, fmt.Errorf("cover locality boundary with H3 cells: %w", err)
	}
	return cells, nil
}

func (r *LocalityRepository) Create(ctx context.Context, l locality.Locality, cells []string) error {
	if len(cells) == 0 {
		return fmt.Errorf("locality requires derived H3 coverage")
	}
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin locality creation: %w", err)
	}
	defer tx.Rollback()
	_, err = tx.ExecContext(ctx, localityCreate, l.ID, l.AdministrativeRegionID, l.Name, l.Slug, l.BoundaryWKT, string(l.Status))
	if err != nil {
		return fmt.Errorf("insert locality: %w", err)
	}
	if err := insertLocalityCells(ctx, tx, l.ID, cells, r.h3.Resolution()); err != nil {
		return err
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit locality creation: %w", err)
	}
	return nil
}

func (r *LocalityRepository) UpdateBoundary(ctx context.Context, id uuid.UUID, boundaryWKT string, cells []string) error {
	if len(cells) == 0 {
		return fmt.Errorf("locality requires derived H3 coverage")
	}
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin locality boundary update: %w", err)
	}
	defer tx.Rollback()
	result, err := tx.ExecContext(ctx, localityUpdateBoundary, boundaryWKT, id)
	if err != nil {
		return fmt.Errorf("update locality boundary: %w", err)
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("check locality boundary update: %w", err)
	}
	if rows != 1 {
		return sql.ErrNoRows
	}
	if _, err := tx.ExecContext(ctx, localityDeleteH3Cells, id); err != nil {
		return fmt.Errorf("replace locality H3 coverage: remove old cells: %w", err)
	}
	if err := insertLocalityCells(ctx, tx, id, cells, r.h3.Resolution()); err != nil {
		return err
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit locality boundary update: %w", err)
	}
	return nil
}

func insertLocalityCells(ctx context.Context, tx *sql.Tx, localityID uuid.UUID, cells []string, resolution int) error {
	if len(cells) == 0 {
		return fmt.Errorf("locality requires derived H3 coverage")
	}
	_, err := tx.ExecContext(ctx, localityInsertH3Cells, localityID, pq.Array(cells), resolution)
	if err != nil {
		return fmt.Errorf("insert locality H3 coverage: %w", err)
	}
	return nil
}

func (r *LocalityRepository) SetStatus(ctx context.Context, id uuid.UUID, status locality.Status) error {
	if status != locality.StatusDraft && status != locality.StatusActive && status != locality.StatusInactive {
		return locality.ErrInvalidStatus
	}
	query := localityUpdateStatus
	if status == locality.StatusActive {
		query = localityActivate
		result, err := r.db.ExecContext(ctx, query, string(status), id, r.h3.Resolution())
		if err != nil {
			return fmt.Errorf("activate locality: %w", err)
		}
		rows, err := result.RowsAffected()
		if err != nil {
			return fmt.Errorf("check locality activation: %w", err)
		}
		if rows != 1 {
			return fmt.Errorf("activate locality: %w", ports.ErrLocalityUnresolved)
		}
		return nil
	}
	result, err := r.db.ExecContext(ctx, query, string(status), id)
	if err != nil {
		return fmt.Errorf("update locality status: %w", err)
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("check locality status update: %w", err)
	}
	if rows != 1 {
		return sql.ErrNoRows
	}
	return nil
}

func (r *LocalityRepository) ResolveH3Cell(ctx context.Context, cell string) (uuid.UUID, error) {
	rows, err := r.db.QueryContext(ctx, localityResolveH3Cell, cell, r.h3.Resolution())
	if err != nil {
		return uuid.Nil, fmt.Errorf("resolve active locality for H3 cell: %w", err)
	}
	defer rows.Close()
	if !rows.Next() {
		if err := rows.Err(); err != nil {
			return uuid.Nil, fmt.Errorf("iterate locality matches: %w", err)
		}
		return uuid.Nil, ports.ErrLocalityUnresolved
	}
	var match uuid.UUID
	if err := rows.Scan(&match); err != nil {
		return uuid.Nil, fmt.Errorf("scan locality match: %w", err)
	}
	if rows.Next() {
		return uuid.Nil, ports.ErrAmbiguousLocality
	}
	if err := rows.Err(); err != nil {
		return uuid.Nil, fmt.Errorf("iterate locality matches: %w", err)
	}
	return match, nil
}

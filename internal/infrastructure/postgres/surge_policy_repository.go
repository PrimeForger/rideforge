package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/ashadashraf/ride-hail-app/internal/domain/surgepolicy"
	"github.com/ashadashraf/ride-hail-app/internal/ports"
	"github.com/google/uuid"
)

type SurgePolicyRepository struct{ db *sql.DB }

func NewSurgePolicyRepository(db *sql.DB) *SurgePolicyRepository {
	return &SurgePolicyRepository{db: db}
}

var _ ports.SurgePolicyRepository = (*SurgePolicyRepository)(nil)

func (r *SurgePolicyRepository) LoadActive(ctx context.Context, now time.Time) ([]surgepolicy.Policy, error) {
	rows, err := r.db.QueryContext(ctx, surgePolicyLoadActive, now)
	if err != nil {
		return nil, fmt.Errorf("query active surge policies: %w", err)
	}
	defer rows.Close()

	policies := make([]surgepolicy.Policy, 0)
	indexes := make(map[uuid.UUID]int)
	for rows.Next() {
		var p surgepolicy.Policy
		var status string
		var until sql.NullTime
		var threshold, multiplier sql.NullFloat64
		if err := rows.Scan(&p.ID, &p.LocalityID, &p.Version, &status,
			&p.MinMultiplier, &p.MaxMultiplier, &p.MinimumDemand,
			&p.EffectiveFrom, &until, &threshold, &multiplier); err != nil {
			return nil, fmt.Errorf("scan active surge policy: %w", err)
		}
		p.Status = surgepolicy.Status(status)
		if until.Valid {
			value := until.Time
			p.EffectiveUntil = &value
		}
		idx, exists := indexes[p.ID]
		if !exists {
			idx = len(policies)
			indexes[p.ID] = idx
			policies = append(policies, p)
		}
		if threshold.Valid && multiplier.Valid {
			policies[idx].RatioTiers = append(policies[idx].RatioTiers, surgepolicy.RatioTier{
				MinImbalanceRatio: threshold.Float64,
				Multiplier:        multiplier.Float64,
			})
		} else if threshold.Valid != multiplier.Valid {
			return nil, fmt.Errorf("surge policy %s has incomplete ratio tier", p.ID)
		}
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate active surge policies: %w", err)
	}
	return policies, nil
}

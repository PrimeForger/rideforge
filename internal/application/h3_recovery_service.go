package application

import (
	"context"
	"fmt"
	"time"

	"github.com/ashadashraf/ride-hail-app/internal/domain/driver"
	"github.com/ashadashraf/ride-hail-app/internal/ports"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type RecoveryResult struct {
	TotalDrivers    int
	RestoredDrivers int
	SkippedDrivers  int
	CellUpdates     int
	Duration        time.Duration
}

type H3RecoveryService struct {
	driverRepo  ports.DriverRepository
	h3          ports.H3CellCalculator
	h3Index     ports.H3DriverIndexer
	geo         ports.GeoIndexer
	driverCache ports.RealtimeDriverCache
	log         *zap.Logger
}

func NewH3RecoveryService(
	driverRepo ports.DriverRepository,
	h3 ports.H3CellCalculator,
	h3Index ports.H3DriverIndexer,
	geo ports.GeoIndexer,
	driverCache ports.RealtimeDriverCache,
	log *zap.Logger,
) *H3RecoveryService {
	return &H3RecoveryService{
		driverRepo:  driverRepo,
		h3:          h3,
		h3Index:     h3Index,
		geo:         geo,
		driverCache: driverCache,
		log:         log,
	}
}

func (s *H3RecoveryService) RebuildIndex(ctx context.Context) (RecoveryResult, error) {
	startTime := time.Now()

	if s.log != nil {
		s.log.Info("H3 driver index recovery starting")
	}

	result := RecoveryResult{}

	// Step 1: Query authoritative available drivers from PostgreSQL repository.
	// PostgreSQL is authoritative for driver existence and eligibility. Failure here fails recovery.
	authDrivers, err := s.driverRepo.GetAvailableDrivers(ctx)
	if err != nil {
		if s.log != nil {
			s.log.Error("failed to query available drivers for recovery", zap.Error(err))
		}
		return RecoveryResult{}, fmt.Errorf("failed to fetch available drivers from repository: %w", err)
	}

	validDriverMap := make(map[uuid.UUID]*driver.Driver)
	invalidDrivers := make([]*driver.Driver, 0)

	for _, d := range authDrivers {
		if d == nil {
			continue
		}
		if d.IsAvailable() {
			validDriverMap[d.ID] = d
		} else {
			invalidDrivers = append(invalidDrivers, d)
		}
	}

	// Step 2: DriverCache may provide fresher coordinates ONLY for drivers in validDriverMap.
	// DriverCache failure is logged at WARN level and recovery continues with PostgreSQL coordinates.
	if s.driverCache != nil {
		cacheOnlineIDs, err := s.driverCache.GetOnlineDriverIDs(ctx)
		if err != nil {
			if s.log != nil {
				s.log.Warn("failed to fetch online driver IDs from driver cache during recovery", zap.Error(err))
			}
		} else if len(cacheOnlineIDs) > 0 {
			targetIDs := make([]uuid.UUID, 0, len(cacheOnlineIDs))
			for _, id := range cacheOnlineIDs {
				if _, exists := validDriverMap[id]; exists {
					targetIDs = append(targetIDs, id)
				}
			}

			if len(targetIDs) > 0 {
				cachedDrivers, err := s.driverCache.LoadDrivers(ctx, targetIDs)
				if err != nil {
					if s.log != nil {
						s.log.Warn("failed to load driver details from driver cache during recovery", zap.Error(err))
					}
				} else {
					for _, cd := range cachedDrivers {
						if cd != nil && isValidLocation(cd.Lat, cd.Lng) {
							if existing, found := validDriverMap[cd.ID]; found {
								existing.Lat = cd.Lat
								existing.Lng = cd.Lng
							}
						}
					}
				}
			}
		}
	}

	// Step 3: Handle PostgreSQL drivers classified as NOT RESTORABLE (ineligible or offline)
	for _, d := range invalidDrivers {
		result.TotalDrivers++
		result.SkippedDrivers++

		if err := s.removeDriverFromDerivedIndexes(ctx, d.ID); err != nil {
			return RecoveryResult{}, err
		}
	}

	// Step 4: Rebuild H3 and Geo index for each valid driver
	restoredSet := make(map[uuid.UUID]struct{}, len(validDriverMap))
	for _, d := range validDriverMap {
		result.TotalDrivers++

		if !isValidLocation(d.Lat, d.Lng) {
			result.SkippedDrivers++
			if err := s.removeDriverFromDerivedIndexes(ctx, d.ID); err != nil {
				return RecoveryResult{}, err
			}
			continue
		}

		cell, err := s.h3.CellForLocation(d.Lat, d.Lng)
		if err != nil {
			if s.log != nil {
				s.log.Warn("failed to calculate H3 cell for driver",
					zap.String("driver_id", d.ID.String()),
					zap.Float64("lat", d.Lat),
					zap.Float64("lng", d.Lng),
					zap.Error(err),
				)
			}
			result.SkippedDrivers++
			if err := s.removeDriverFromDerivedIndexes(ctx, d.ID); err != nil {
				return RecoveryResult{}, err
			}
			continue
		}

		updateRes, err := s.h3Index.UpdateDriverCell(ctx, d.ID, cell)
		if err != nil {
			if s.log != nil {
				s.log.Error("failed to update H3 driver cell during recovery",
					zap.String("driver_id", d.ID.String()),
					zap.String("cell", cell),
					zap.Error(err),
				)
			}
			return RecoveryResult{}, fmt.Errorf("failed to update H3 driver cell for driver %s: %w", d.ID, err)
		}

		if updateRes.Status != ports.DriverCellUnchanged {
			result.CellUpdates++
		}

		if s.geo != nil {
			if err := s.geo.UpdateDriverLocation(ctx, d.ID, d.Lat, d.Lng); err != nil {
				if s.log != nil {
					s.log.Error("failed to update geo driver location during recovery",
						zap.String("driver_id", d.ID.String()),
						zap.Error(err),
					)
				}
				return RecoveryResult{}, fmt.Errorf("failed to update geo location for driver %s: %w", d.ID, err)
			}
		}

		result.RestoredDrivers++
		restoredSet[d.ID] = struct{}{}
	}

	// Step 5: Reconcile & remove stale drivers present in Redis online set but missing from restored valid driver set
	if s.driverCache != nil {
		redisOnlineIDs, err := s.driverCache.GetOnlineDriverIDs(ctx)
		if err != nil {
			if s.log != nil {
				s.log.Warn("failed to fetch online driver IDs for stale driver cleanup", zap.Error(err))
			}
		} else {
			for _, redisID := range redisOnlineIDs {
				if _, active := restoredSet[redisID]; !active {
					if err := s.removeDriverFromDerivedIndexes(ctx, redisID); err != nil {
						return RecoveryResult{}, err
					}
				}
			}
		}
	}

	result.Duration = time.Since(startTime)

	if s.log != nil {
		s.log.Info("H3 driver index recovery completed",
			zap.Int("total_drivers", result.TotalDrivers),
			zap.Int("restored_drivers", result.RestoredDrivers),
			zap.Int("skipped_drivers", result.SkippedDrivers),
			zap.Int("cell_updates", result.CellUpdates),
			zap.Duration("duration", result.Duration),
		)
	}

	return result, nil
}

func (s *H3RecoveryService) removeDriverFromDerivedIndexes(ctx context.Context, driverID uuid.UUID) error {
	if s.h3Index != nil {
		if _, err := s.h3Index.RemoveDriver(ctx, driverID); err != nil {
			if s.log != nil {
				s.log.Error("failed to remove stale driver from H3 index",
					zap.String("driver_id", driverID.String()),
					zap.Error(err),
				)
			}
			return fmt.Errorf("failed to remove stale H3 driver cell for driver %s: %w", driverID, err)
		}
	}

	if s.geo != nil {
		if err := s.geo.RemoveDriver(ctx, driverID); err != nil {
			if s.log != nil {
				s.log.Error("failed to remove stale driver from geo index",
					zap.String("driver_id", driverID.String()),
					zap.Error(err),
				)
			}
			return fmt.Errorf("failed to remove stale geo driver location for driver %s: %w", driverID, err)
		}
	}

	return nil
}

func isValidLocation(lat, lng float64) bool {
	if lat < -90 || lat > 90 || lng < -180 || lng > 180 {
		return false
	}
	if lat == 0 && lng == 0 {
		return false
	}
	return true
}

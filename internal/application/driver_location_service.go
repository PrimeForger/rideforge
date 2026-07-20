package application

import (
	"context"

	"github.com/ashadashraf/ride-hail-app/internal/config"
	appgeo "github.com/ashadashraf/ride-hail-app/internal/infrastructure/geo"
	"github.com/ashadashraf/ride-hail-app/internal/infrastructure/redis"
	"github.com/google/uuid"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
)

type DriverLocationService struct {
	geo         *redis.GeoService
	driverCache *redis.DriverCache
	h3          *appgeo.H3Service
	h3Index     *redis.H3DriverIndex
	cfg         *config.Config
}

func NewDriverLocationService(
	geo *redis.GeoService,
	driverCache *redis.DriverCache,
	h3 *appgeo.H3Service,
	h3Index *redis.H3DriverIndex,
	cfg *config.Config,
) *DriverLocationService {
	return &DriverLocationService{
		geo:         geo,
		driverCache: driverCache,
		h3:          h3,
		h3Index:     h3Index,
		cfg:         cfg,
	}
}

var driverLocationTracer = otel.Tracer("application.driver_location")
var spatialTracer = otel.Tracer("application.spatial")

func (s *DriverLocationService) ProcessRealtimeLocation(
	ctx context.Context,
	driverID uuid.UUID,
	lat, lng float64,
	accuracy float64,
	speed float64,
	bearing float64,
	seq int64,
) error {

	ctx, span := driverLocationTracer.Start(ctx, "driver.location.process_realtime")
	defer span.End()

	span.SetAttributes(
		attribute.String("driver.id", driverID.String()),
		// attribute.Float64("location.lat", lat),
		// attribute.Float64("location.lng", lng),
		// attribute.Float64("location.accuracy", accuracy),
		// attribute.Float64("location.speed", speed),
		// attribute.Float64("location.bearing", bearing),
		attribute.Int64("location.seq", seq),
	)

	if lat < -90 || lat > 90 || lng < -180 || lng > 180 {
		span.SetAttributes(attribute.String("location.reject_reason", "invalid_coordinates"))
		span.SetStatus(codes.Ok, "rejected_invalid_coordinates")
		return ErrInvalidCoordinates
	}

	if accuracy <= 0 || accuracy > s.cfg.Realtime.MaxLocationAccuracyMeters {
		span.SetAttributes(attribute.String("location.reject_reason", "bad_accuracy"))
		span.SetStatus(codes.Ok, "rejected_bad_accuracy")
		return ErrBadAccuracy
	}

	if seq <= 0 {
		span.SetAttributes(attribute.String("location.reject_reason", "invalid_sequence"))
		span.SetStatus(codes.Ok, "rejected_invalid_sequence")
		return ErrInvalidSequence
	}

	accepted, err := s.driverCache.AcceptLocationSeq(ctx, driverID, seq)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "accept_location_sequence_failed")
		return err
	}

	if !accepted {
		span.SetAttributes(attribute.String("location.reject_reason", "stale_sequence"))
		span.SetStatus(codes.Ok, "rejected_stale_sequence")
		return ErrStaleSequence
	}

	if err := s.updateSpatialIndexes(ctx, driverID, lat, lng); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "update_spatial_indexes_failed")
		return err
	}

	if err := s.driverCache.UpdateDriverLocationDetails(
		ctx,
		driverID,
		lat,
		lng,
		accuracy,
		speed,
		bearing,
		seq,
	); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "update_driver_location_details_failed")
		return err
	}

	if err := s.driverCache.RefreshHeartbeat(ctx, driverID); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "refresh_driver_heartbeat_failed")
		return err
	}

	if err := s.driverCache.RefreshConnection(ctx, driverID); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "refresh_driver_connection_failed")
		return err
	}

	span.SetAttributes(
		attribute.Bool("location.persisted", true),
	)

	span.SetStatus(codes.Ok, "location_processed")

	return nil
}

func (s *DriverLocationService) PersistLocation(
	ctx context.Context,
	driverID uuid.UUID,
	lat, lng float64,
) error {

	ctx, span := driverLocationTracer.Start(ctx, "driver.location.persist")
	defer span.End()

	span.SetAttributes(
		attribute.String("driver.id", driverID.String()),
		attribute.Float64("location.lat", lat),
		attribute.Float64("location.lng", lng),
	)

	if lat < -90 || lat > 90 || lng < -180 || lng > 180 {
		span.SetAttributes(attribute.String("location.reject_reason", "invalid_coordinates"))
		span.SetStatus(codes.Ok, "rejected_invalid_coordinates")
		return ErrInvalidCoordinates
	}

	if err := s.updateSpatialIndexes(ctx, driverID, lat, lng); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "update_spatial_indexes_failed")
		return err
	}

	if err := s.driverCache.UpdateDriverLocation(
		ctx,
		driverID,
		lat,
		lng,
	); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "update_driver_location_cache_failed")
		return err
	}

	span.SetAttributes(
		attribute.Bool("location.persisted", true),
	)

	span.SetStatus(codes.Ok, "location_persisted")

	return nil
}

func (s *DriverLocationService) updateSpatialIndexes(
	ctx context.Context,
	driverID uuid.UUID,
	lat, lng float64,
) error {

	ctx, span := spatialTracer.Start(ctx, "driver.spatial_update")
	defer span.End()

	span.SetAttributes(
		attribute.String("driver.id", driverID.String()),
		attribute.Bool("h3.enabled", s.cfg.H3.Enabled),
	)

	ctx, geoSpan := spatialTracer.Start(ctx, "redis.geo.update")

	err := s.geo.UpdateDriverLocation(ctx, driverID, lat, lng)

	if err != nil {
		geoSpan.RecordError(err)
		geoSpan.SetStatus(codes.Error, "geo_update_failed")
		geoSpan.End()
		return err
	}

	geoSpan.SetStatus(codes.Ok, "geo_updated")
	geoSpan.End()

	if !s.cfg.H3.Enabled {
		return nil
	}

	ctx, h3Span := spatialTracer.Start(ctx, "h3.cell.update")
	defer h3Span.End()

	cell, err := s.h3.CellForLocation(lat, lng)
	if err != nil {
		return err
	}

	updateResult, err := s.h3Index.UpdateDriverCell(ctx, driverID, cell)

	if err != nil {
		h3Span.RecordError(err)
		h3Span.SetStatus(codes.Error, "h3_update_failed")
		return err
	}

	h3Span.SetAttributes(
		attribute.String("h3.cell", cell),
	)

	switch updateResult.Status {

	case redis.DriverCellAdded:
		h3Span.SetAttributes(
			attribute.String("h3.update_status", "added"),
		)

	case redis.DriverCellMoved:
		h3Span.SetAttributes(
			attribute.String("h3.update_status", "moved"),
		)

	case redis.DriverCellUnchanged:
		h3Span.SetAttributes(
			attribute.String("h3.update_status", "unchanged"),
		)
	}

	return err
}

package redis

import (
	"context"

	"github.com/ashadashraf/ride-hail-app/internal/domain/driver"
	"github.com/redis/go-redis/v9"

	"github.com/google/uuid"
)

const driverGeoKey = "drivers:geo"

type GeoService struct {
	client *Client
}

func NewGeoService(client *Client) *GeoService {
	return &GeoService{client: client}
}

func (g *GeoService) UpdateDriverLocation(
	ctx context.Context,
	driverID uuid.UUID,
	lat, lng float64,
) error {

	rdb := g.client.GetRaw()

	return rdb.GeoAdd(ctx, driverGeoKey, &redis.GeoLocation{
		Name:      driverID.String(),
		Latitude:  lat,
		Longitude: lng,
	}).Err()
}

func (g *GeoService) SetDriverDetails(
	ctx context.Context,
	d *driver.Driver,
) error {

	key := "driver:" + d.ID.String()

	return g.client.rdb.HSet(ctx, key, map[string]interface{}{
		"status":            string(d.Status),
		"rating":            d.Rating,
		"acceptance_rate":   d.AcceptanceRate,
		"cancellation_rate": d.CancellationRate,
		"timeout_rate":      d.TimeoutRate,
		"completed_rides":   d.CompletedRides,
		"last_assigned_at":  d.LastAssignedAt.Unix(),
		"lat":               d.Lat,
		"lng":               d.Lng,
	}).Err()
}

func (g *GeoService) FindNearbyDriversWithDistance(
	ctx context.Context,
	lat, lng float64,
	radiusKm float64,
	limit int,
) ([]struct {
	ID       uuid.UUID
	Distance float64
}, error) {

	rdb := g.client.GetRaw()

	results, err := rdb.GeoSearchLocation(
		ctx,
		driverGeoKey,
		&redis.GeoSearchLocationQuery{
			GeoSearchQuery: redis.GeoSearchQuery{
				Longitude:  lng,
				Latitude:   lat,
				Radius:     radiusKm,
				RadiusUnit: "km",
				Sort:       "ASC",
				Count:      limit,
			},
			WithDist: true,
		},
	).Result()

	if err != nil {
		return nil, err
	}

	var drivers []struct {
		ID       uuid.UUID
		Distance float64
	}

	for _, res := range results {

		id, err := uuid.Parse(res.Name)
		if err != nil {
			continue
		}

		drivers = append(drivers, struct {
			ID       uuid.UUID
			Distance float64
		}{
			ID:       id,
			Distance: res.Dist,
		})
	}

	return drivers, nil
}

func (g *GeoService) RemoveDriver(
	ctx context.Context,
	driverID uuid.UUID,
) error {

	rdb := g.client.GetRaw()

	return rdb.ZRem(ctx, driverGeoKey, driverID.String()).Err()
}

func (g *GeoService) Distance(
	ctx context.Context,
	lat, lng float64,
	driverID uuid.UUID,
) float64 {

	res, err := g.client.rdb.GeoDist(
		ctx,
		driverGeoKey,
		driverID.String(),
		driverID.String(), // TEMP: replace with point vs driver
		"km",
	).Result()

	if err != nil {
		return 5 // fallback
	}

	return res
}

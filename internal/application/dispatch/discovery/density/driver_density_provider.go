package density

import "context"

type DriverDensityProvider interface {
	DriverCountInRing(ctx context.Context, centerCell string, ring int) (int, error)
}

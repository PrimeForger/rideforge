package memory

import (
	"context"
	"sync"

	"github.com/ashadashraf/ride-hail-app/internal/domain/ride"
	"github.com/google/uuid"
)

type RideRepository struct {
	store map[uuid.UUID]*ride.Ride
	mu    sync.RWMutex
}

func NewRideRepository() *RideRepository {
	return &RideRepository{
		store: make(map[uuid.UUID]*ride.Ride),
	}
}

func (r *RideRepository) Save(ctx context.Context, entity *ride.Ride) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.store[entity.ID] = entity
	return nil
}

func (r *RideRepository) GetByID(ctx context.Context, id uuid.UUID) (*ride.Ride, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	return r.store[id], nil
}

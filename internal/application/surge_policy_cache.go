package application

import (
	"context"
	"fmt"
	"sync/atomic"
	"time"

	"github.com/ashadashraf/ride-hail-app/internal/domain/surgepolicy"
	"github.com/ashadashraf/ride-hail-app/internal/ports"
	"github.com/google/uuid"
)

type policySnapshot struct {
	policies map[uuid.UUID]surgepolicy.Policy
	loadedAt time.Time
}

type SurgePolicyCache struct {
	repo         ports.SurgePolicyRepository
	maxStaleness time.Duration
	now          func() time.Time
	snapshot     atomic.Pointer[policySnapshot]
	lastSuccess  atomic.Int64
	lastFailure  atomic.Int64
}

func NewSurgePolicyCache(repo ports.SurgePolicyRepository, maxStaleness time.Duration) *SurgePolicyCache {
	return &SurgePolicyCache{repo: repo, maxStaleness: maxStaleness, now: time.Now}
}

func (c *SurgePolicyCache) Refresh(ctx context.Context) error {
	now := c.now()
	recordFailure := func(err error) error {
		c.lastFailure.Store(now.UnixNano())
		return err
	}
	policies, err := c.repo.LoadActive(ctx, now)
	if err != nil {
		return recordFailure(fmt.Errorf("load surge policies: %w", err))
	}
	for _, p := range policies {
		if err := p.Validate(); err != nil {
			return recordFailure(fmt.Errorf("invalid active surge policy %s: %w", p.ID, err))
		}
	}
	if err := surgepolicy.ValidateNoOverlappingActivePolicies(policies); err != nil {
		return recordFailure(fmt.Errorf("validate active surge policy schedule: %w", err))
	}
	next := make(map[uuid.UUID]surgepolicy.Policy, len(policies))
	for _, p := range policies {
		if !p.ValidAt(now) {
			continue
		}
		if _, exists := next[p.LocalityID]; exists {
			return recordFailure(fmt.Errorf("duplicate active policy for locality %s", p.LocalityID))
		}
		p.RatioTiers = append([]surgepolicy.RatioTier(nil), p.RatioTiers...)
		next[p.LocalityID] = p
	}
	c.snapshot.Store(&policySnapshot{policies: next, loadedAt: now})
	c.lastSuccess.Store(now.UnixNano())
	return nil
}

func (c *SurgePolicyCache) PolicyForLocality(localityID string) (surgepolicy.Policy, bool) {
	id, err := uuid.Parse(localityID)
	if err != nil {
		return surgepolicy.Policy{}, false
	}
	s := c.snapshot.Load()
	if s == nil {
		return surgepolicy.Policy{}, false
	}
	if c.maxStaleness <= 0 || c.now().Sub(s.loadedAt) > c.maxStaleness {
		return surgepolicy.Policy{}, false
	}
	p, ok := s.policies[id]
	p.RatioTiers = append([]surgepolicy.RatioTier(nil), p.RatioTiers...)
	return p, ok
}

func (c *SurgePolicyCache) SnapshotState() (int, time.Time, time.Time) {
	s := c.snapshot.Load()
	entries := 0
	if s != nil {
		entries = len(s.policies)
	}
	return entries, time.Unix(0, c.lastSuccess.Load()), time.Unix(0, c.lastFailure.Load())
}

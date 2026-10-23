package surge

import (
	"context"
	"errors"
	"github.com/ashadashraf/ride-hail-app/internal/domain/surgepolicy"
	"github.com/google/uuid"
	"testing"
	"time"
)

type policyRepoStub struct {
	policies []surgepolicy.Policy
	err      error
}

func (r *policyRepoStub) LoadActive(context.Context, time.Time) ([]surgepolicy.Policy, error) {
	return r.policies, r.err
}

func TestSurgePolicyCacheRetainsSnapshotOnRefreshFailure(t *testing.T) {
	id := uuid.New()
	now := time.Now()
	r := &policyRepoStub{policies: []surgepolicy.Policy{{ID: uuid.New(), LocalityID: id, Version: 1, Status: surgepolicy.StatusActive, MinMultiplier: 1, MaxMultiplier: 1.5, EffectiveFrom: now.Add(-time.Hour), RatioTiers: []surgepolicy.RatioTier{{MinImbalanceRatio: 1, Multiplier: 1.1}}}}}
	c := NewSurgePolicyCache(r, time.Hour)
	c.now = func() time.Time { return now }
	if err := c.Refresh(context.Background()); err != nil {
		t.Fatal(err)
	}
	_, lastSuccessBefore, _ := c.SnapshotState()
	now = now.Add(time.Minute)
	c.now = func() time.Time { return now }
	r.err = errors.New("database unavailable")
	if err := c.Refresh(context.Background()); err == nil {
		t.Fatal("expected refresh failure")
	}
	if _, ok := c.PolicyForLocality(id.String()); !ok {
		t.Fatal("previous snapshot was discarded")
	}
	_, lastSuccessAfter, lastFailure := c.SnapshotState()
	if !lastSuccessAfter.Equal(lastSuccessBefore) {
		t.Fatalf("failed refresh advanced last success: before=%v after=%v", lastSuccessBefore, lastSuccessAfter)
	}
	if !lastFailure.Equal(now) {
		t.Fatalf("last failure = %v, want %v", lastFailure, now)
	}
}

func TestSurgePolicyCacheRejectsStaleSnapshot(t *testing.T) {
	id := uuid.New()
	now := time.Now()
	r := &policyRepoStub{policies: []surgepolicy.Policy{{ID: uuid.New(), LocalityID: id, Version: 1, Status: surgepolicy.StatusActive, MinMultiplier: 1, MaxMultiplier: 1.5, EffectiveFrom: now.Add(-time.Hour), RatioTiers: []surgepolicy.RatioTier{{MinImbalanceRatio: 1, Multiplier: 1.1}}}}}
	c := NewSurgePolicyCache(r, time.Minute)
	c.now = func() time.Time { return now }
	if err := c.Refresh(context.Background()); err != nil {
		t.Fatal(err)
	}
	r.err = errors.New("database unavailable")
	c.now = func() time.Time { return now.Add(30 * time.Second) }
	if err := c.Refresh(context.Background()); err == nil {
		t.Fatal("expected refresh failure")
	}
	c.now = func() time.Time { return now.Add(2 * time.Minute) }
	if _, ok := c.PolicyForLocality(id.String()); ok {
		t.Fatal("stale policy returned")
	}
}

func TestSurgePolicyCacheInvalidAndDuplicateRefreshRetainSnapshot(t *testing.T) {
	id := uuid.New()
	now := time.Now()
	r := &policyRepoStub{policies: []surgepolicy.Policy{{
		ID: uuid.New(), LocalityID: id, Version: 1, Status: surgepolicy.StatusActive,
		MinMultiplier: 1, MaxMultiplier: 1.5, EffectiveFrom: now.Add(-time.Hour),
		RatioTiers: []surgepolicy.RatioTier{{MinImbalanceRatio: 1, Multiplier: 1.1}},
	}}}
	c := NewSurgePolicyCache(r, time.Hour)
	c.now = func() time.Time { return now }
	if err := c.Refresh(context.Background()); err != nil {
		t.Fatal(err)
	}
	_, lastSuccessBefore, _ := c.SnapshotState()
	now = now.Add(time.Minute)
	c.now = func() time.Time { return now }
	if _, ok := c.PolicyForLocality(id.String()); !ok {
		t.Fatal("valid policy missing before invalid refresh")
	}
	r.policies[0].RatioTiers = nil
	if err := c.Refresh(context.Background()); err == nil {
		t.Fatal("invalid policy refresh unexpectedly succeeded")
	}
	if policy, ok := c.PolicyForLocality(id.String()); !ok || policy.MaxMultiplier != 1.5 {
		t.Fatalf("invalid refresh did not retain last known-good policy: %+v, available=%t", policy, ok)
	}
	_, lastSuccessAfter, lastFailure := c.SnapshotState()
	if !lastSuccessAfter.Equal(lastSuccessBefore) {
		t.Fatalf("invalid refresh advanced last success: before=%v after=%v", lastSuccessBefore, lastSuccessAfter)
	}
	if !lastFailure.Equal(now) {
		t.Fatalf("last failure = %v, want %v", lastFailure, now)
	}
	r.policies[0].RatioTiers = []surgepolicy.RatioTier{{MinImbalanceRatio: 1, Multiplier: 1.1}}
	r.policies = append(r.policies, r.policies[0])
	if err := c.Refresh(context.Background()); err == nil {
		t.Fatal("duplicate effective policy refresh unexpectedly succeeded")
	}
	_, duplicateLastSuccess, duplicateLastFailure := c.SnapshotState()
	if !duplicateLastSuccess.Equal(lastSuccessBefore) || !duplicateLastFailure.Equal(now) {
		t.Fatalf("duplicate refresh timestamps: success=%v failure=%v; want success=%v failure=%v", duplicateLastSuccess, duplicateLastFailure, lastSuccessBefore, now)
	}
	if policy, ok := c.PolicyForLocality(id.String()); !ok || policy.Version != 1 {
		t.Fatalf("duplicate refresh did not retain last known-good policy: %+v, available=%t", policy, ok)
	}
}

func TestSurgePolicyCacheWithoutSuccessfulSnapshotIsUnavailable(t *testing.T) {
	id := uuid.New()
	r := &policyRepoStub{err: errors.New("database unavailable")}
	c := NewSurgePolicyCache(r, time.Hour)
	if _, ok := c.PolicyForLocality(id.String()); ok {
		t.Fatal("policy available before successful refresh")
	}
	if err := c.Refresh(context.Background()); err == nil {
		t.Fatal("expected initial refresh failure")
	}
	if _, ok := c.PolicyForLocality(id.String()); ok {
		t.Fatal("policy available after no successful refresh")
	}
}

func TestSurgePolicyCacheReturnsIndependentTierSlice(t *testing.T) {
	id := uuid.New()
	now := time.Now()
	r := &policyRepoStub{policies: []surgepolicy.Policy{{
		ID: uuid.New(), LocalityID: id, Version: 1, Status: surgepolicy.StatusActive,
		MinMultiplier: 1, MaxMultiplier: 1.5, EffectiveFrom: now.Add(-time.Hour),
		RatioTiers: []surgepolicy.RatioTier{{MinImbalanceRatio: 1.2, Multiplier: 1.1}},
	}}}
	c := NewSurgePolicyCache(r, time.Hour)
	c.now = func() time.Time { return now }
	if err := c.Refresh(context.Background()); err != nil {
		t.Fatal(err)
	}
	policy, ok := c.PolicyForLocality(id.String())
	if !ok {
		t.Fatal("policy missing")
	}
	policy.RatioTiers[0].Multiplier = 1.5
	policy, _ = c.PolicyForLocality(id.String())
	if policy.RatioTiers[0].Multiplier != 1.1 {
		t.Fatal("caller mutation changed the cached policy snapshot")
	}
}

package profile_test

import (
	"reflect"
	"testing"

	candidatedensity "github.com/ashadashraf/ride-hail-app/internal/application/dispatch/discovery/density"
	"github.com/ashadashraf/ride-hail-app/internal/application/dispatch/discovery/profile"
	"github.com/ashadashraf/ride-hail-app/internal/application/dispatch/discovery/search"
)

func TestDefaultExpansionProfileProvider_Profile(t *testing.T) {
	testCfg := profile.ExpansionConfig{
		Sparse: profile.ExpansionProfile{
			RingStep:      3,
			MaxRing:       7,
			MaxCells:      222,
			MaxCandidates: 333,
		},
		Normal: profile.ExpansionProfile{
			RingStep:      2,
			MaxRing:       5,
			MaxCells:      111,
			MaxCandidates: 155,
		},
		Dense: profile.ExpansionProfile{
			RingStep:      1,
			MaxRing:       2,
			MaxCells:      45,
			MaxCandidates: 50,
		},
		Default: profile.ExpansionProfile{
			RingStep:      1,
			MaxRing:       4,
			MaxCells:      90,
			MaxCandidates: 95,
		},
	}

	provider := profile.NewDefaultExpansionProfileProvider(testCfg)

	tests := []struct {
		name        string
		density     candidatedensity.DensityClass
		wantProfile profile.ExpansionProfile
	}{
		{
			name:        "sparse density returns configured sparse profile",
			density:     candidatedensity.DensitySparse,
			wantProfile: testCfg.Sparse,
		},
		{
			name:        "normal density returns configured normal profile",
			density:     candidatedensity.DensityNormal,
			wantProfile: testCfg.Normal,
		},
		{
			name:        "dense density returns configured dense profile",
			density:     candidatedensity.DensityDense,
			wantProfile: testCfg.Dense,
		},
		{
			name:        "unknown density returns configured default profile",
			density:     candidatedensity.DensityUnknown,
			wantProfile: testCfg.Default,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			got := provider.Profile(tt.density)
			if !reflect.DeepEqual(got, tt.wantProfile) {
				t.Errorf("Profile(%v) = %+v, want %+v", tt.density, got, tt.wantProfile)
			}
		})
	}
}

func TestDefaultExpansionPolicy(t *testing.T) {
	p := profile.ExpansionProfile{
		RingStep:      2,
		MaxRing:       3,
		MaxCells:      50,
		MaxCandidates: 50,
	}

	policy := profile.NewDefaultExpansionPolicy(p)

	budget := search.NewSearchBudget(100, 1)
	policy.ConfigureBudget(&budget)

	if budget.MaxRing != 3 || budget.MaxCells != 50 || budget.CandidateLimit != 50 {
		t.Errorf("ConfigureBudget failed: %+v", budget)
	}

	// Test NextRing within bounds
	if next := policy.NextRing(0); next != 2 {
		t.Errorf("expected NextRing(0)=2, got %d", next)
	}

	// Test NextRing clamped to MaxRing
	if next := policy.NextRing(2); next != 3 {
		t.Errorf("expected NextRing(2)=3 (clamped), got %d", next)
	}
}

func TestProfileBuilder(t *testing.T) {
	builder := profile.NewBuilder()
	builder.SetName("custom-profile")

	expPolicy := profile.NewDefaultExpansionPolicy(profile.ExpansionProfile{
		RingStep: 1,
		MaxRing:  2,
	})
	builder.SetExpansion(expPolicy)

	built := builder.Build()
	if built.Name() != "custom-profile" {
		t.Errorf("expected name 'custom-profile', got %q", built.Name())
	}
	if built.Expansion() != expPolicy {
		t.Errorf("expected expansion policy to match")
	}
}

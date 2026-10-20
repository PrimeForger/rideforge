package config

import (
	"testing"
	"time"
)

func TestLoadDatabaseSettings(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://db.example/rideforge")
	t.Setenv("DATABASE_CONNECT_TIMEOUT", "-1s")
	cfg := Load()
	if cfg.DatabaseURL != "postgres://db.example/rideforge" {
		t.Fatalf("database URL was not loaded")
	}
	if cfg.DatabaseConnectTimeout != 5*time.Second {
		t.Fatalf("invalid connect timeout = %s, want 5s", cfg.DatabaseConnectTimeout)
	}

	t.Setenv("DATABASE_CONNECT_TIMEOUT", "2s")
	cfg = Load()
	if cfg.DatabaseConnectTimeout != 2*time.Second {
		t.Fatalf("configured connect timeout = %s, want 2s", cfg.DatabaseConnectTimeout)
	}
}

func TestConfig_ValidateH3_NonPositiveInterval(t *testing.T) {
	cfg := &Config{
		H3: H3Config{
			ReconciliationIntervalSeconds: 0,
		},
	}
	cfg.validateH3()

	if cfg.H3.ReconciliationIntervalSeconds != 300 {
		t.Errorf("expected ReconciliationIntervalSeconds to default to 300, got %d", cfg.H3.ReconciliationIntervalSeconds)
	}

	cfgNegative := &Config{
		H3: H3Config{
			ReconciliationIntervalSeconds: -10,
		},
	}
	cfgNegative.validateH3()

	if cfgNegative.H3.ReconciliationIntervalSeconds != 300 {
		t.Errorf("expected ReconciliationIntervalSeconds to default to 300 for negative input, got %d", cfgNegative.H3.ReconciliationIntervalSeconds)
	}
}

func TestValidateSurgePolicyKeepsBoundedPositiveRefresh(t *testing.T) {
	cfg := &Config{SurgePolicy: SurgePolicyConfig{RefreshIntervalSeconds: 0, MaxStalenessSeconds: 0}}
	cfg.validateSurgePolicy()
	if cfg.SurgePolicy.RefreshIntervalSeconds <= 0 || cfg.SurgePolicy.MaxStalenessSeconds <= cfg.SurgePolicy.RefreshIntervalSeconds {
		t.Fatalf("invalid normalized surge refresh config: %+v", cfg.SurgePolicy)
	}
	cfg = &Config{SurgePolicy: SurgePolicyConfig{RefreshIntervalSeconds: 20, MaxStalenessSeconds: 10}}
	cfg.validateSurgePolicy()
	if cfg.SurgePolicy.MaxStalenessSeconds <= cfg.SurgePolicy.RefreshIntervalSeconds {
		t.Fatalf("staleness must exceed refresh interval: %+v", cfg.SurgePolicy)
	}
}

func TestConfig_ValidateDemand_NonPositiveInterval(t *testing.T) {
	for _, interval := range []int{0, -10} {
		cfg := &Config{Demand: DemandConfig{ReconciliationIntervalSeconds: interval}}
		cfg.validateDemand()
		if cfg.Demand.ReconciliationIntervalSeconds != 300 {
			t.Errorf("interval %d normalized to %d, want 300", interval, cfg.Demand.ReconciliationIntervalSeconds)
		}
	}
}

func TestConfig_Load_NormalizesDemandReconciliationInterval(t *testing.T) {
	for _, interval := range []string{"0", "-5"} {
		t.Run(interval, func(t *testing.T) {
			t.Setenv("DEMAND_RECONCILIATION_INTERVAL_SECONDS", interval)
			cfg := Load()
			if cfg.Demand.ReconciliationIntervalSeconds <= 0 {
				t.Fatalf("demand reconciliation interval must be positive, got %d", cfg.Demand.ReconciliationIntervalSeconds)
			}
		})
	}
}

func TestConfig_ExpansionDefaults(t *testing.T) {
	cfg := Load()

	if cfg.Expansion.SparseRingStep != 2 {
		t.Errorf("expected SparseRingStep=2, got %d", cfg.Expansion.SparseRingStep)
	}
	if cfg.Expansion.SparseMaxCells != 150 {
		t.Errorf("expected SparseMaxCells=150, got %d", cfg.Expansion.SparseMaxCells)
	}
	if cfg.Expansion.SparseMaxCandidates != 150 {
		t.Errorf("expected SparseMaxCandidates=150, got %d", cfg.Expansion.SparseMaxCandidates)
	}

	if cfg.Expansion.NormalRingStep != 1 {
		t.Errorf("expected NormalRingStep=1, got %d", cfg.Expansion.NormalRingStep)
	}
	if cfg.Expansion.NormalMaxCells != 80 {
		t.Errorf("expected NormalMaxCells=80, got %d", cfg.Expansion.NormalMaxCells)
	}
	if cfg.Expansion.NormalMaxCandidates != 80 {
		t.Errorf("expected NormalMaxCandidates=80, got %d", cfg.Expansion.NormalMaxCandidates)
	}

	if cfg.Expansion.DenseRingStep != 1 {
		t.Errorf("expected DenseRingStep=1, got %d", cfg.Expansion.DenseRingStep)
	}
	if cfg.Expansion.DenseMaxRing != 1 {
		t.Errorf("expected DenseMaxRing=1, got %d", cfg.Expansion.DenseMaxRing)
	}
	if cfg.Expansion.DenseMaxCells != 30 {
		t.Errorf("expected DenseMaxCells=30, got %d", cfg.Expansion.DenseMaxCells)
	}
	if cfg.Expansion.DenseMaxCandidates != 30 {
		t.Errorf("expected DenseMaxCandidates=30, got %d", cfg.Expansion.DenseMaxCandidates)
	}

	if cfg.Expansion.DefaultRingStep != 1 {
		t.Errorf("expected DefaultRingStep=1, got %d", cfg.Expansion.DefaultRingStep)
	}
	if cfg.Expansion.DefaultMaxCells != 60 {
		t.Errorf("expected DefaultMaxCells=60, got %d", cfg.Expansion.DefaultMaxCells)
	}
	if cfg.Expansion.DefaultMaxCandidates != 60 {
		t.Errorf("expected DefaultMaxCandidates=60, got %d", cfg.Expansion.DefaultMaxCandidates)
	}
}

func TestConfig_ExpansionEnvOverrides(t *testing.T) {
	t.Setenv("EXPANSION_SPARSE_RING_STEP", "4")
	t.Setenv("EXPANSION_SPARSE_MAX_CELLS", "300")
	t.Setenv("EXPANSION_SPARSE_MAX_CANDIDATES", "350")
	t.Setenv("EXPANSION_NORMAL_RING_STEP", "2")
	t.Setenv("EXPANSION_NORMAL_MAX_CELLS", "120")
	t.Setenv("EXPANSION_NORMAL_MAX_CANDIDATES", "130")
	t.Setenv("EXPANSION_DENSE_RING_STEP", "3")
	t.Setenv("EXPANSION_DENSE_MAX_RING", "2")
	t.Setenv("EXPANSION_DENSE_MAX_CELLS", "50")
	t.Setenv("EXPANSION_DENSE_MAX_CANDIDATES", "55")
	t.Setenv("EXPANSION_DEFAULT_RING_STEP", "2")
	t.Setenv("EXPANSION_DEFAULT_MAX_CELLS", "90")
	t.Setenv("EXPANSION_DEFAULT_MAX_CANDIDATES", "95")

	cfg := Load()

	if cfg.Expansion.SparseRingStep != 4 {
		t.Errorf("expected SparseRingStep=4, got %d", cfg.Expansion.SparseRingStep)
	}
	if cfg.Expansion.SparseMaxCells != 300 {
		t.Errorf("expected SparseMaxCells=300, got %d", cfg.Expansion.SparseMaxCells)
	}
	if cfg.Expansion.SparseMaxCandidates != 350 {
		t.Errorf("expected SparseMaxCandidates=350, got %d", cfg.Expansion.SparseMaxCandidates)
	}

	if cfg.Expansion.NormalRingStep != 2 {
		t.Errorf("expected NormalRingStep=2, got %d", cfg.Expansion.NormalRingStep)
	}
	if cfg.Expansion.NormalMaxCells != 120 {
		t.Errorf("expected NormalMaxCells=120, got %d", cfg.Expansion.NormalMaxCells)
	}
	if cfg.Expansion.NormalMaxCandidates != 130 {
		t.Errorf("expected NormalMaxCandidates=130, got %d", cfg.Expansion.NormalMaxCandidates)
	}

	if cfg.Expansion.DenseRingStep != 3 {
		t.Errorf("expected DenseRingStep=3, got %d", cfg.Expansion.DenseRingStep)
	}
	if cfg.Expansion.DenseMaxRing != 2 {
		t.Errorf("expected DenseMaxRing=2, got %d", cfg.Expansion.DenseMaxRing)
	}
	if cfg.Expansion.DenseMaxCells != 50 {
		t.Errorf("expected DenseMaxCells=50, got %d", cfg.Expansion.DenseMaxCells)
	}
	if cfg.Expansion.DenseMaxCandidates != 55 {
		t.Errorf("expected DenseMaxCandidates=55, got %d", cfg.Expansion.DenseMaxCandidates)
	}

	if cfg.Expansion.DefaultRingStep != 2 {
		t.Errorf("expected DefaultRingStep=2, got %d", cfg.Expansion.DefaultRingStep)
	}
	if cfg.Expansion.DefaultMaxCells != 90 {
		t.Errorf("expected DefaultMaxCells=90, got %d", cfg.Expansion.DefaultMaxCells)
	}
	if cfg.Expansion.DefaultMaxCandidates != 95 {
		t.Errorf("expected DefaultMaxCandidates=95, got %d", cfg.Expansion.DefaultMaxCandidates)
	}
}

func TestConfig_NormalizeWeights(t *testing.T) {
	cfg := &Config{
		Ranking: RankingConfig{
			DistanceWeight:     10,
			AcceptanceWeight:   10,
			CancellationWeight: 10,
			TimeoutWeight:      10,
			RatingWeight:       10,
			ExperienceWeight:   25,
			FairnessWeight:     25,
		},
	}
	cfg.normalizeWeights()

	total := cfg.Ranking.DistanceWeight +
		cfg.Ranking.AcceptanceWeight +
		cfg.Ranking.CancellationWeight +
		cfg.Ranking.TimeoutWeight +
		cfg.Ranking.RatingWeight +
		cfg.Ranking.ExperienceWeight +
		cfg.Ranking.FairnessWeight

	if total < 0.999 || total > 1.001 {
		t.Errorf("expected normalized weights sum ~ 1.0, got %f", total)
	}

	// Zero weights branch
	cfgZero := &Config{}
	cfgZero.normalizeWeights()
}

func TestConfig_ValidateMatching(t *testing.T) {
	cfgValid := &Config{
		Matching: MatchingConfig{
			SparseDriverThreshold: 5,
			DenseDriverThreshold:  20,
		},
	}
	cfgValid.validateMatching()

	defer func() {
		if r := recover(); r == nil {
			t.Errorf("expected panic when sparse >= dense")
		}
	}()

	cfgInvalid := &Config{
		Matching: MatchingConfig{
			SparseDriverThreshold: 30,
			DenseDriverThreshold:  20,
		},
	}
	cfgInvalid.validateMatching()
}

func TestConfig_HelperParsers(t *testing.T) {
	t.Setenv("TEST_INT_VALID", "42")
	t.Setenv("TEST_INT_INVALID", "abc")
	if v := getInt("TEST_INT_VALID", 1); v != 42 {
		t.Errorf("expected 42, got %d", v)
	}
	if v := getInt("TEST_INT_INVALID", 1); v != 1 {
		t.Errorf("expected fallback 1, got %d", v)
	}
	if v := getInt("TEST_INT_UNSET", 99); v != 99 {
		t.Errorf("expected fallback 99, got %d", v)
	}

	t.Setenv("TEST_FLOAT_VALID", "3.14")
	t.Setenv("TEST_FLOAT_INVALID", "xyz")
	if v := getFloat("TEST_FLOAT_VALID", 1.0); v != 3.14 {
		t.Errorf("expected 3.14, got %f", v)
	}
	if v := getFloat("TEST_FLOAT_INVALID", 1.0); v != 1.0 {
		t.Errorf("expected fallback 1.0, got %f", v)
	}
	if v := getFloat("TEST_FLOAT_UNSET", 2.5); v != 2.5 {
		t.Errorf("expected fallback 2.5, got %f", v)
	}

	t.Setenv("TEST_STR", "val")
	if v := getString("TEST_STR", "def"); v != "val" {
		t.Errorf("expected val, got %s", v)
	}
	if v := getString("TEST_STR_UNSET", "def"); v != "def" {
		t.Errorf("expected def, got %s", v)
	}

	t.Setenv("TEST_BOOL_VALID", "true")
	t.Setenv("TEST_BOOL_INVALID", "notabool")
	if v := getBool("TEST_BOOL_VALID", false); !v {
		t.Errorf("expected true, got %t", v)
	}
	if v := getBool("TEST_BOOL_INVALID", true); !v {
		t.Errorf("expected fallback true, got %t", v)
	}
	if v := getBool("TEST_BOOL_UNSET", true); !v {
		t.Errorf("expected fallback true, got %t", v)
	}

	t.Setenv("TEST_DUR_VALID", "5s")
	t.Setenv("TEST_DUR_INVALID", "invalid")
	if v := getDuration("TEST_DUR_VALID", 0); v.Seconds() != 5 {
		t.Errorf("expected 5s, got %v", v)
	}
	if v := getDuration("TEST_DUR_INVALID", 0); v != 0 {
		t.Errorf("expected fallback 0, got %v", v)
	}
	if v := getDuration("TEST_DUR_UNSET", 10); v != 10 {
		t.Errorf("expected fallback 10, got %v", v)
	}
}

package config

import (
	"testing"
)

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

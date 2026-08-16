package builders_test

import (
	"testing"

	"github.com/ashadashraf/ride-hail-app/internal/bootstrap/builders"
	"github.com/ashadashraf/ride-hail-app/internal/config"
)

func TestBuildRoutingProvider_Supported(t *testing.T) {
	tests := []struct {
		provider string
	}{
		{provider: "none"},
		{provider: ""},
	}

	for _, tt := range tests {
		cfg := &config.Config{
			Routing: config.RoutingConfig{
				Provider: tt.provider,
			},
		}

		provider, err := builders.BuildRoutingProvider(cfg)
		if err != nil {
			t.Errorf("expected no error for provider %q, got %v", tt.provider, err)
		}
		if provider == nil {
			t.Errorf("expected non-nil provider for %q", tt.provider)
		}
	}
}

func TestBuildRoutingProvider_Unsupported(t *testing.T) {
	cfg := &config.Config{
		Routing: config.RoutingConfig{
			Provider: "unknown_provider_xyz",
		},
	}

	provider, err := builders.BuildRoutingProvider(cfg)
	if err == nil {
		t.Fatalf("expected error for unsupported provider, got nil")
	}
	if provider != nil {
		t.Errorf("expected nil provider on error, got %v", provider)
	}
}

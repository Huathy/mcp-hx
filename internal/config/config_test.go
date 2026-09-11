package config

import (
	"testing"
	"time"
)

func TestEffectiveSafetyInheritsGlobalQueryTimeout(t *testing.T) {
	cfg := &Config{
		Safety: SafetyConfig{
			Mode:         "read-write",
			MaxRows:      1000,
			QueryTimeout: Duration(30 * time.Second),
		},
		DataSources: []DataSourceConfig{
			{
				Name: "main-mysql",
				Safety: &SafetyConfig{
					Mode: "read-write",
				},
			},
		},
	}
	got := cfg.EffectiveSafety("main-mysql")
	if got.QueryTimeout.Std() != 30*time.Second {
		t.Errorf("QueryTimeout = %v, want 30s (should inherit global)", got.QueryTimeout.Std())
	}
	if got.MaxRows != 1000 {
		t.Errorf("MaxRows = %d, want 1000 (should inherit global)", got.MaxRows)
	}
}

func TestEffectiveSafetyOverridesGlobal(t *testing.T) {
	cfg := &Config{
		Safety: SafetyConfig{
			Mode:         "read-write",
			MaxRows:      1000,
			QueryTimeout: Duration(30 * time.Second),
		},
		DataSources: []DataSourceConfig{
			{
				Name: "main-mysql",
				Safety: &SafetyConfig{
					Mode:         "read-only",
					MaxRows:      500,
					QueryTimeout: Duration(10 * time.Second),
				},
			},
		},
	}
	got := cfg.EffectiveSafety("main-mysql")
	if got.Mode != "read-only" {
		t.Errorf("Mode = %s, want read-only (ds override)", got.Mode)
	}
	if got.MaxRows != 500 {
		t.Errorf("MaxRows = %d, want 500 (ds override)", got.MaxRows)
	}
	if got.QueryTimeout.Std() != 10*time.Second {
		t.Errorf("QueryTimeout = %v, want 10s (ds override)", got.QueryTimeout.Std())
	}
}

func TestEffectiveSafetyNoDSConfig(t *testing.T) {
	cfg := &Config{
		Safety: SafetyConfig{
			Mode:         "read-write",
			MaxRows:      1000,
			QueryTimeout: Duration(30 * time.Second),
		},
		DataSources: []DataSourceConfig{
			{Name: "main-mysql"},
		},
	}
	got := cfg.EffectiveSafety("main-mysql")
	if got.QueryTimeout.Std() != 30*time.Second {
		t.Errorf("QueryTimeout = %v, want 30s (global fallback)", got.QueryTimeout.Std())
	}
}

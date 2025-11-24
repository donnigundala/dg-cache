package memory

import (
	"testing"
	"time"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	if cfg.MaxItems != 0 {
		t.Errorf("Expected MaxItems to be 0, got %d", cfg.MaxItems)
	}

	if cfg.MaxBytes != 0 {
		t.Errorf("Expected MaxBytes to be 0, got %d", cfg.MaxBytes)
	}

	if cfg.EvictionPolicy != "lru" {
		t.Errorf("Expected EvictionPolicy to be 'lru', got %s", cfg.EvictionPolicy)
	}

	if cfg.CleanupInterval != 1*time.Minute {
		t.Errorf("Expected CleanupInterval to be 1 minute, got %v", cfg.CleanupInterval)
	}

	if cfg.EnableMetrics {
		t.Error("Expected EnableMetrics to be false")
	}
}

func TestConfigBuilders(t *testing.T) {
	cfg := DefaultConfig().
		WithMaxItems(1000).
		WithMaxBytes(10 * 1024 * 1024).
		WithEvictionPolicy("lfu").
		WithCleanupInterval(30 * time.Second).
		WithMetrics(true)

	if cfg.MaxItems != 1000 {
		t.Errorf("Expected MaxItems to be 1000, got %d", cfg.MaxItems)
	}

	if cfg.MaxBytes != 10*1024*1024 {
		t.Errorf("Expected MaxBytes to be 10MB, got %d", cfg.MaxBytes)
	}

	if cfg.EvictionPolicy != "lfu" {
		t.Errorf("Expected EvictionPolicy to be 'lfu', got %s", cfg.EvictionPolicy)
	}

	if cfg.CleanupInterval != 30*time.Second {
		t.Errorf("Expected CleanupInterval to be 30 seconds, got %v", cfg.CleanupInterval)
	}

	if !cfg.EnableMetrics {
		t.Error("Expected EnableMetrics to be true")
	}
}

func TestConfigImmutability(t *testing.T) {
	cfg1 := DefaultConfig()
	cfg2 := cfg1.WithMaxItems(100)

	if cfg1.MaxItems != 0 {
		t.Error("Original config should not be modified")
	}

	if cfg2.MaxItems != 100 {
		t.Error("New config should have updated value")
	}
}

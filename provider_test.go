package cache

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCacheServiceProvider_Name(t *testing.T) {
	provider := &CacheServiceProvider{}
	assert.Equal(t, "cache", provider.Name())
}

func TestCacheServiceProvider_Version(t *testing.T) {
	provider := &CacheServiceProvider{}
	assert.Equal(t, "1.6.2", provider.Version())
}

func TestCacheServiceProvider_Dependencies(t *testing.T) {
	provider := &CacheServiceProvider{}
	deps := provider.Dependencies()

	assert.NotNil(t, deps)
	assert.Empty(t, deps, "dg-cache should have no required dependencies")
}

func TestCacheServiceProvider_ConfigDefaults(t *testing.T) {
	provider := &CacheServiceProvider{}
	// Config should use zero values initially
	assert.Equal(t, "", provider.Config.DefaultStore)
}

func TestCacheServiceProvider_CustomConfig(t *testing.T) {
	customConfig := Config{
		DefaultStore: "memory",
		Prefix:       "test",
	}

	provider := &CacheServiceProvider{
		Config: customConfig,
	}

	assert.Equal(t, "memory", provider.Config.DefaultStore)
	assert.Equal(t, "test", provider.Config.Prefix)
}

func TestCacheServiceProvider_DriverFactories(t *testing.T) {
	mockFactory := func(config StoreConfig) (Driver, error) {
		return nil, nil
	}

	provider := &CacheServiceProvider{
		DriverFactories: map[string]DriverFactory{
			"memory": mockFactory,
		},
	}

	assert.NotNil(t, provider.DriverFactories)
	assert.Contains(t, provider.DriverFactories, "memory")
}

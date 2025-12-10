package cache_test

import (
	"context"
	"testing"

	cache "github.com/donnigundala/dg-cache"
	"github.com/donnigundala/dg-cache/drivers/memory"
	"github.com/donnigundala/dg-core/foundation"
	"github.com/stretchr/testify/assert"
)

// -----------------------------------------------------------------------------
// Existing Helper Tests (Testing Manager methods directly)
// -----------------------------------------------------------------------------

func TestManager_GetAs(t *testing.T) {
	manager, _ := cache.NewManager(cache.DefaultConfig())

	type User struct {
		ID   int
		Name string
	}

	ctx := context.Background()
	_ = User{ID: 1, Name: "John"} // Create but don't use (no store to test with)

	// This test would need a real store implementation
	// For now, just verify the method exists and compiles
	var result User
	_ = manager.GetAs(ctx, "user:1", &result)
}

func TestManager_GetString(t *testing.T) {
	manager, _ := cache.NewManager(cache.DefaultConfig())
	ctx := context.Background()

	// Verify method exists
	_, _ = manager.GetString(ctx, "key")
}

func TestManager_GetInt(t *testing.T) {
	manager, _ := cache.NewManager(cache.DefaultConfig())
	ctx := context.Background()

	// Verify method exists
	_, _ = manager.GetInt(ctx, "key")
}

func TestManager_GetInt64(t *testing.T) {
	manager, _ := cache.NewManager(cache.DefaultConfig())
	ctx := context.Background()

	// Verify method exists
	_, _ = manager.GetInt64(ctx, "key")
}

func TestManager_GetFloat64(t *testing.T) {
	manager, _ := cache.NewManager(cache.DefaultConfig())
	ctx := context.Background()

	// Verify method exists
	_, _ = manager.GetFloat64(ctx, "key")
}

func TestManager_GetBool(t *testing.T) {
	manager, _ := cache.NewManager(cache.DefaultConfig())
	ctx := context.Background()

	// Verify method exists
	_, _ = manager.GetBool(ctx, "key")
}

// -----------------------------------------------------------------------------
// Container Integration Tests (v1.6.0)
// -----------------------------------------------------------------------------

func TestResolve(t *testing.T) {
	app := foundation.New(".")
	config := cache.DefaultConfig()

	provider := &cache.CacheServiceProvider{Config: config}
	err := provider.Register(app)
	assert.NoError(t, err)

	// Test Resolve
	c, err := cache.Resolve(app)
	assert.NoError(t, err)
	assert.NotNil(t, c)
}

func TestResolve_Error(t *testing.T) {
	app := foundation.New(".")

	// Test Resolve without registration
	c, err := cache.Resolve(app)
	assert.Error(t, err)
	assert.Nil(t, c)
	assert.Contains(t, err.Error(), "failed to resolve cache")
}

func TestMustResolve(t *testing.T) {
	app := foundation.New(".")
	config := cache.DefaultConfig()

	provider := &cache.CacheServiceProvider{Config: config}
	err := provider.Register(app)
	assert.NoError(t, err)

	// Test MustResolve
	c := cache.MustResolve(app)
	assert.NotNil(t, c)
}

func TestMustResolve_Panic(t *testing.T) {
	app := foundation.New(".")

	// Test MustResolve panics without registration
	assert.Panics(t, func() {
		cache.MustResolve(app)
	})
}

func TestResolveStore(t *testing.T) {
	app := foundation.New(".")
	config := cache.DefaultConfig()

	// Add named store
	config = config.WithStore("redis", cache.StoreConfig{
		Driver: "memory", // Use memory driver for test
	})

	provider := &cache.CacheServiceProvider{
		Config: config,
		DriverFactories: map[string]cache.DriverFactory{
			"memory": memory.NewDriver,
		},
	}
	err := provider.Register(app)
	assert.NoError(t, err)

	// Test ResolveStore
	store, err := cache.ResolveStore(app, "redis")
	assert.NoError(t, err)
	assert.NotNil(t, store)
}

func TestResolveStore_Error(t *testing.T) {
	app := foundation.New(".")

	// Test ResolveStore without registration
	store, err := cache.ResolveStore(app, "redis")
	assert.Error(t, err)
	assert.Nil(t, store)
	assert.Contains(t, err.Error(), "failed to resolve cache store")
}

func TestMustResolveStore(t *testing.T) {
	app := foundation.New(".")
	config := cache.DefaultConfig()
	config = config.WithStore("redis", cache.StoreConfig{
		Driver: "memory",
	})

	provider := &cache.CacheServiceProvider{
		Config: config,
		DriverFactories: map[string]cache.DriverFactory{
			"memory": memory.NewDriver,
		},
	}
	err := provider.Register(app)
	assert.NoError(t, err)

	// Test MustResolveStore
	store := cache.MustResolveStore(app, "redis")
	assert.NotNil(t, store)
}

func TestMustResolveStore_Panic(t *testing.T) {
	app := foundation.New(".")

	// Test MustResolveStore panics without registration
	assert.Panics(t, func() {
		cache.MustResolveStore(app, "redis")
	})
}

func TestInjectable(t *testing.T) {
	app := foundation.New(".")
	config := cache.DefaultConfig()
	config = config.WithStore("redis", cache.StoreConfig{
		Driver: "memory",
	})

	provider := &cache.CacheServiceProvider{
		Config: config,
		DriverFactories: map[string]cache.DriverFactory{
			"memory": memory.NewDriver,
		},
	}
	err := provider.Register(app)
	assert.NoError(t, err)

	// Test Injectable
	inject := cache.NewInjectable(app)
	assert.NotNil(t, inject)

	// Test Cache()
	c := inject.Cache()
	assert.NotNil(t, c)

	// Test Store()
	store := inject.Store("redis")
	assert.NotNil(t, store)

	// Test TryStore() - existing
	tryStore := inject.TryStore("redis")
	assert.NotNil(t, tryStore)

	// Test TryStore() - non-existing
	nilStore := inject.TryStore("nonexistent")
	assert.Nil(t, nilStore)
}

func TestInjectable_Panic(t *testing.T) {
	app := foundation.New(".")

	inject := cache.NewInjectable(app)

	// Test Cache() panics without registration
	assert.Panics(t, func() {
		inject.Cache()
	})

	// Test Store() panics without registration
	assert.Panics(t, func() {
		inject.Store("redis")
	})
}

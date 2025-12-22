package dgcache_test

import (
	"context"
	"testing"
	"time"

	cache "github.com/donnigundala/dg-cache"
	dgcache "github.com/donnigundala/dg-cache"
	"github.com/donnigundala/dg-cache/drivers/memory"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func createManager(t *testing.T) *dgcache.Manager {
	cfg := dgcache.DefaultConfig()
	manager, err := dgcache.NewManager(cfg)
	require.NoError(t, err)

	// Register memory driver
	manager.RegisterDriver("memory", memory.NewDriver)

	return manager
}

func TestManager_BasicOperations(t *testing.T) {
	manager := createManager(t)
	ctx := context.Background()

	// Put
	err := manager.Put(ctx, "key1", "value1", 1*time.Minute)
	assert.NoError(t, err)

	// Get
	val, err := manager.Get(ctx, "key1")
	assert.NoError(t, err)
	assert.Equal(t, "value1", val)

	// Has
	has, err := manager.Has(ctx, "key1")
	assert.NoError(t, err)
	assert.True(t, has)

	// Missing
	missing, err := manager.Missing(ctx, "key2")
	assert.NoError(t, err)
	assert.True(t, missing)

	// Forget
	err = manager.Forget(ctx, "key1")
	assert.NoError(t, err)

	has, err = manager.Has(ctx, "key1")
	assert.NoError(t, err)
	assert.False(t, has)
}

func TestManager_TTL(t *testing.T) {
	manager := createManager(t)
	ctx := context.Background()

	// Put with short TTL
	err := manager.Put(ctx, "short", "value", 100*time.Millisecond)
	assert.NoError(t, err)

	// Should exist immediately
	val, err := manager.Get(ctx, "short")
	assert.NoError(t, err)
	assert.Equal(t, "value", val)

	// Wait for expiration
	time.Sleep(200 * time.Millisecond)

	// Should be gone
	val, err = manager.Get(ctx, "short")
	assert.Equal(t, dgcache.ErrKeyNotFound, err)
	assert.Nil(t, val)
}

func TestManager_IncrementDecrement(t *testing.T) {
	manager := createManager(t)
	ctx := context.Background()

	// Increment new key
	val, err := manager.Increment(ctx, "counter", 1)
	assert.NoError(t, err)
	assert.Equal(t, int64(1), val)

	// Increment existing
	val, err = manager.Increment(ctx, "counter", 2)
	assert.NoError(t, err)
	assert.Equal(t, int64(3), val)

	// Decrement
	val, err = manager.Decrement(ctx, "counter", 1)
	assert.NoError(t, err)
	assert.Equal(t, int64(2), val)
}

func TestManager_Remember(t *testing.T) {
	manager := createManager(t)
	ctx := context.Background()
	called := 0

	callback := func() (interface{}, error) {
		called++
		return "computed", nil
	}

	// First call - should execute callback
	val, err := manager.Remember(ctx, "rem_key", 1*time.Minute, callback)
	assert.NoError(t, err)
	assert.Equal(t, "computed", val)
	assert.Equal(t, 1, called)

	// Second call - should return cached value
	val, err = manager.Remember(ctx, "rem_key", 1*time.Minute, callback)
	assert.NoError(t, err)
	assert.Equal(t, "computed", val)
	assert.Equal(t, 1, called) // Callback count should not increase
}

func TestManager_Pull(t *testing.T) {
	manager := createManager(t)
	ctx := context.Background()

	err := manager.Put(ctx, "pull_key", "value", 1*time.Minute)
	assert.NoError(t, err)

	// Pull
	val, err := manager.Pull(ctx, "pull_key")
	assert.NoError(t, err)
	assert.Equal(t, "value", val)

	// Should be gone
	has, err := manager.Has(ctx, "pull_key")
	assert.NoError(t, err)
	assert.False(t, has)
}

func TestManager_Flush(t *testing.T) {
	manager := createManager(t)
	ctx := context.Background()

	manager.Put(ctx, "k1", "v1", 1*time.Minute)
	manager.Put(ctx, "k2", "v2", 1*time.Minute)

	err := manager.Flush(ctx)
	assert.NoError(t, err)

	has1, _ := manager.Has(ctx, "k1")
	has2, _ := manager.Has(ctx, "k2")
	assert.False(t, has1)
	assert.False(t, has2)
}

func TestManager_MultipleStores(t *testing.T) {
	cfg := cache.DefaultConfig()
	cfg = cfg.WithStore("secondary", cache.StoreConfig{
		Driver: "memory",
		Prefix: "sec",
	})

	manager, err := cache.NewManager(cfg)
	require.NoError(t, err)
	manager.RegisterDriver("memory", memory.NewDriver)

	ctx := context.Background()

	// Put in default store
	err = manager.Put(ctx, "key", "default_val", 1*time.Minute)
	assert.NoError(t, err)

	// Put in secondary store
	store, err := manager.Store("secondary")
	assert.NoError(t, err)
	err = store.Put(ctx, "key", "sec_val", 1*time.Minute)
	assert.NoError(t, err)

	// Verify separation
	val1, _ := manager.Get(ctx, "key")
	assert.Equal(t, "default_val", val1)

	val2, _ := store.Get(ctx, "key")
	assert.Equal(t, "sec_val", val2)
}

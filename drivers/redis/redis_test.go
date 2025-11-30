package redis_test

import (
	"context"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	cache "github.com/donnigundala/dg-cache"
	driver "github.com/donnigundala/dg-cache/drivers/redis"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func createDriver(t *testing.T) (cache.Driver, *miniredis.Miniredis) {
	s, err := miniredis.Run()
	require.NoError(t, err)

	addr := s.Addr()
	parts := strings.Split(addr, ":")
	port, _ := strconv.Atoi(parts[1])

	cfg := cache.StoreConfig{
		Driver: "redis",
		Prefix: "test",
		Options: map[string]interface{}{
			"host": parts[0],
			"port": port,
		},
	}

	d, err := driver.NewDriver(cfg)
	require.NoError(t, err)

	return d, s
}

func TestRedis_Configuration(t *testing.T) {
	s, err := miniredis.Run()
	require.NoError(t, err)
	defer s.Close()

	addr := s.Addr()
	parts := strings.Split(addr, ":")
	port, _ := strconv.Atoi(parts[1])

	cfg := cache.StoreConfig{
		Driver: "redis",
		Prefix: "test",
		Options: map[string]interface{}{
			"host":      parts[0],
			"port":      port,
			"database":  1,
			"pool_size": 20,
		},
	}

	d, err := driver.NewDriver(cfg)
	require.NoError(t, err)
	defer d.Close()

	// Verify connection works
	err = d.Put(context.Background(), "test", "value", 1*time.Minute)
	assert.NoError(t, err)
}

func TestRedis_BasicOperations(t *testing.T) {
	d, s := createDriver(t)
	defer s.Close()
	defer d.Close()

	ctx := context.Background()

	// Put
	err := d.Put(ctx, "key1", "value1", 1*time.Minute)
	assert.NoError(t, err)

	// Get
	val, err := d.Get(ctx, "key1")
	assert.NoError(t, err)
	assert.Equal(t, "value1", val)

	// Has
	has, err := d.Has(ctx, "key1")
	assert.NoError(t, err)
	assert.True(t, has)

	// Missing
	missing, err := d.Missing(ctx, "key2")
	assert.NoError(t, err)
	assert.True(t, missing)

	// Forget
	err = d.Forget(ctx, "key1")
	assert.NoError(t, err)

	has, err = d.Has(ctx, "key1")
	assert.NoError(t, err)
	assert.False(t, has)
}

func TestRedis_TTL(t *testing.T) {
	d, s := createDriver(t)
	defer s.Close()
	defer d.Close()

	ctx := context.Background()

	// Put with TTL
	err := d.Put(ctx, "ttl_key", "value", 1*time.Second)
	assert.NoError(t, err)

	// Fast forward time
	s.FastForward(2 * time.Second)

	// Should be gone
	val, err := d.Get(ctx, "ttl_key")
	assert.Equal(t, cache.ErrKeyNotFound, err)
	assert.Nil(t, val)
}

func TestRedis_IncrementDecrement(t *testing.T) {
	d, s := createDriver(t)
	defer s.Close()
	defer d.Close()

	ctx := context.Background()

	// Increment
	val, err := d.Increment(ctx, "counter", 1)
	assert.NoError(t, err)
	assert.Equal(t, int64(1), val)

	val, err = d.Increment(ctx, "counter", 2)
	assert.NoError(t, err)
	assert.Equal(t, int64(3), val)

	// Decrement
	val, err = d.Decrement(ctx, "counter", 1)
	assert.NoError(t, err)
	assert.Equal(t, int64(2), val)
}

func TestRedis_TaggedCache(t *testing.T) {
	d, s := createDriver(t)
	defer s.Close()
	defer d.Close()

	// We need to cast to TaggedStore to use Tags
	// But wait, the Driver struct doesn't implement Tags directly,
	// it implements the method that returns TaggedStore.
	// But Driver interface in dgcore-cache doesn't have Tags() method.
	// Only TaggedStore interface has Tags().
	// However, our redis.Driver struct has a Tags() method.
	// So we need to type assert.

	// Wait, in my implementation of redis.go, I defined:
	// func (d *Driver) Tags(tags ...string) cache.TaggedStore

	// So I can cast d to interface{ Tags(...string) cache.TaggedStore }

	type Taggable interface {
		Tags(tags ...string) cache.TaggedStore
	}

	taggable, ok := d.(Taggable)
	require.True(t, ok, "Driver should implement Tags()")

	ctx := context.Background()

	// Put tagged item
	tagged := taggable.Tags("users", "premium")
	err := tagged.Put(ctx, "user:1", "data", 1*time.Minute)
	assert.NoError(t, err)

	// Verify it exists via normal get
	val, err := d.Get(ctx, "user:1")
	assert.NoError(t, err)
	assert.Equal(t, "data", val)

	// Flush tags
	err = tagged.FlushTags(ctx, "premium")
	assert.NoError(t, err)

	// Verify it's gone
	exists, err := d.Has(ctx, "user:1")
	assert.NoError(t, err)
	assert.False(t, exists)
}

func TestRedis_MultipleTags(t *testing.T) {
	d, s := createDriver(t)
	defer s.Close()
	defer d.Close()

	type Taggable interface {
		Tags(tags ...string) cache.TaggedStore
	}
	taggable := d.(Taggable)
	ctx := context.Background()

	// Item with tag1
	taggable.Tags("tag1").Put(ctx, "k1", "v1", 1*time.Minute)

	// Item with tag2
	taggable.Tags("tag2").Put(ctx, "k2", "v2", 1*time.Minute)

	// Item with both
	taggable.Tags("tag1", "tag2").Put(ctx, "k3", "v3", 1*time.Minute)

	// Flush tag1
	taggable.Tags("tag1").FlushTags(ctx)

	// k1 and k3 should be gone
	has1, _ := d.Has(ctx, "k1")
	has2, _ := d.Has(ctx, "k2")
	has3, _ := d.Has(ctx, "k3")

	assert.False(t, has1, "k1 should be deleted")
	assert.True(t, has2, "k2 should remain")
	assert.False(t, has3, "k3 should be deleted")
}

func TestRedis_GetMultiple(t *testing.T) {
	d, s := createDriver(t)
	defer s.Close()
	defer d.Close()

	ctx := context.Background()

	// Setup data
	d.Put(ctx, "k1", "v1", 1*time.Minute)
	d.Put(ctx, "k2", "v2", 1*time.Minute)

	// Get multiple
	vals, err := d.GetMultiple(ctx, []string{"k1", "k2", "k3"})
	assert.NoError(t, err)
	assert.Len(t, vals, 2)
	assert.Equal(t, "v1", vals["k1"])
	assert.Equal(t, "v2", vals["k2"])
	assert.NotContains(t, vals, "k3")
}

func TestRedis_PutMultiple(t *testing.T) {
	d, s := createDriver(t)
	defer s.Close()
	defer d.Close()

	ctx := context.Background()

	items := map[string]interface{}{
		"k1": "v1",
		"k2": "v2",
	}

	err := d.PutMultiple(ctx, items, 1*time.Minute)
	assert.NoError(t, err)

	// Verify
	val1, _ := d.Get(ctx, "k1")
	val2, _ := d.Get(ctx, "k2")
	assert.Equal(t, "v1", val1)
	assert.Equal(t, "v2", val2)
}

func TestRedis_Forever(t *testing.T) {
	d, s := createDriver(t)
	defer s.Close()
	defer d.Close()

	ctx := context.Background()

	err := d.Forever(ctx, "forever_key", "value")
	assert.NoError(t, err)

	val, err := d.Get(ctx, "forever_key")
	assert.NoError(t, err)
	assert.Equal(t, "value", val)

	// Check TTL (should be 0 or -1 depending on implementation, but definitely not expiring soon)
	// Miniredis returns 0 for no expiration
	ttl := s.TTL("test:forever_key")
	assert.Equal(t, time.Duration(0), ttl)
}

func TestRedis_Flush(t *testing.T) {
	d, s := createDriver(t)
	defer s.Close()
	defer d.Close()

	ctx := context.Background()

	d.Put(ctx, "k1", "v1", 1*time.Minute)
	d.Put(ctx, "k2", "v2", 1*time.Minute)

	err := d.Flush(ctx)
	assert.NoError(t, err)

	has1, _ := d.Has(ctx, "k1")
	has2, _ := d.Has(ctx, "k2")
	assert.False(t, has1)
	assert.False(t, has2)
}

func TestRedis_GettersSetters(t *testing.T) {
	d, s := createDriver(t)
	defer s.Close()
	defer d.Close()

	assert.Equal(t, "redis", d.Name())
	assert.Equal(t, "test", d.GetPrefix())

	d.SetPrefix("new_prefix")
	assert.Equal(t, "new_prefix", d.GetPrefix())
}

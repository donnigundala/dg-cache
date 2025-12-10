package memory

import (
	"context"
	"testing"
	"time"

	cache "github.com/donnigundala/dg-cache"
	"github.com/stretchr/testify/assert"
)

func TestTaggedCache(t *testing.T) {
	driver, err := NewDriver(cache.StoreConfig{
		Driver: "memory",
	})
	assert.NoError(t, err)

	ctx := context.Background()

	// 1. Put Item with Tags
	err = driver.(cache.TaggedStore).Tags("users", "admins").Put(ctx, "user:1", "john", time.Minute)
	assert.NoError(t, err)

	// 2. Verify Item Exists
	val, err := driver.Get(ctx, "user:1")
	assert.NoError(t, err)
	assert.Equal(t, "john", val)

	// 3. Put another item with different tag
	err = driver.(cache.TaggedStore).Tags("posts").Put(ctx, "post:1", "hollola", time.Minute)
	assert.NoError(t, err)

	// 4. Flush "users" tag
	err = driver.(cache.TaggedStore).Tags("users").Flush(ctx)
	assert.NoError(t, err)

	// 5. Verify user:1 is gone
	exists, err := driver.Has(ctx, "user:1")
	assert.NoError(t, err)
	assert.False(t, exists)

	// 6. Verify post:1 still exists
	exists, err = driver.Has(ctx, "post:1")
	assert.NoError(t, err)
	assert.True(t, exists)

	// 7. Flush "posts" tag
	err = driver.(cache.TaggedStore).Tags("posts").Flush(ctx)
	assert.NoError(t, err)

	// 8. Verify post:1 is gone
	exists, err = driver.Has(ctx, "post:1")
	assert.NoError(t, err)
	assert.False(t, exists)
}

func TestTaggedCache_Cleanup(t *testing.T) {
	driver, err := NewDriver(cache.StoreConfig{})
	assert.NoError(t, err)
	ctx := context.Background()
	memDriver := driver.(*Driver)

	// Put item with tag
	driver.(cache.TaggedStore).Tags("tag1").Put(ctx, "key1", "val1", time.Minute)

	// Verify internal state
	assert.Contains(t, memDriver.tags, "tag1")
	assert.Contains(t, memDriver.tags["tag1"], "key1")

	// Forget item directly
	driver.Forget(ctx, "key1")

	// Verify cleanup
	assert.NotContains(t, memDriver.tags, "tag1")
}

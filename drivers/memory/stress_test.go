package memory

import (
	"context"
	"fmt"
	"math/rand"
	"sync"
	"testing"
	"time"

	dgcache "github.com/donnigundala/dg-cache"
	"github.com/donnigundala/dg-core/contracts/cache"
	"github.com/stretchr/testify/assert"
)

// TestConcurrency_ReadWrite verifies safe concurrent access for Get and Put.
func TestConcurrency_ReadWrite(t *testing.T) {
	driver, err := NewDriver(dgcache.StoreConfig{
		Driver: "memory",
	})
	assert.NoError(t, err)

	ctx := context.Background()
	var wg sync.WaitGroup
	numGoroutines := 50
	numOps := 100

	// Writers
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < numOps; j++ {
				key := fmt.Sprintf("key-%d-%d", id, j)
				err := driver.Put(ctx, key, j, time.Minute)
				assert.NoError(t, err)
			}
		}(i)
	}

	// Readers
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < numOps; j++ {
				// Read keys written by random writers (or self)
				// Here we just prevent panics, don't strictly assert existence as it's racey
				key := fmt.Sprintf("key-%d-%d", rand.Intn(numGoroutines), rand.Intn(numOps))
				_, _ = driver.Get(ctx, key)
			}
		}(i)
	}

	wg.Wait()
}

// TestConcurrency_Tags verifies safe concurrent access for Tagging and Flushing.
// This is critical to catch deadlocks between Put (locking) and FlushTags (locking).
func TestConcurrency_Tags(t *testing.T) {
	driver, err := NewDriver(dgcache.StoreConfig{
		Driver: "memory",
	})
	assert.NoError(t, err)

	ctx := context.Background()
	var wg sync.WaitGroup
	numGoroutines := 20
	numOps := 50

	// Tag Writers
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			tag := fmt.Sprintf("tag-%d", id%5) // Shared tags
			tagged := driver.(cache.TaggedStore).Tags(tag)
			for j := 0; j < numOps; j++ {
				key := fmt.Sprintf("key-%d-%d", id, j)
				err := tagged.Put(ctx, key, "value", time.Minute)
				assert.NoError(t, err)
			}
		}(i)
	}

	// Tag Flushers
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			tag := fmt.Sprintf("tag-%d", id)
			tagged := driver.(cache.TaggedStore).Tags(tag)
			for j := 0; j < numOps/2; j++ {
				err := tagged.Flush(ctx)
				assert.NoError(t, err)
				time.Sleep(time.Millisecond) // Slight delay to allow writes
			}
		}(i)
	}

	wg.Wait()
}

// TestConcurrency_Eviction verifies safe concurrent access under heavy load forcing eviction.
func TestConcurrency_Eviction(t *testing.T) {
	driver, err := NewDriver(dgcache.StoreConfig{
		Options: map[string]interface{}{
			"max_items": 100, // Small limit to force evictions
		},
	})
	assert.NoError(t, err)

	ctx := context.Background()
	var wg sync.WaitGroup
	numGoroutines := 50
	numOps := 200

	// Writers forcing eviction
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < numOps; j++ {
				key := fmt.Sprintf("evict-%d-%d", id, j)
				err := driver.Put(ctx, key, "data", time.Minute)
				assert.NoError(t, err)
			}
		}(i)
	}

	wg.Wait()

	// Verify size didn't explode (approximate)
	memDriver := driver.(*Driver)
	assert.LessOrEqual(t, len(memDriver.items), 100)
}

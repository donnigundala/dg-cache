package memory

import (
	"context"
	"testing"
	"time"

	dgcache "github.com/donnigundala/dg-cache"
)

func TestDriver_MaxItemsEviction(t *testing.T) {
	config := dgcache.StoreConfig{
		Driver: "memory",
		Options: map[string]interface{}{
			"max_items":      3,
			"enable_metrics": true,
		},
	}

	driver, err := NewDriver(config)
	if err != nil {
		t.Fatalf("Failed to create driver: %v", err)
	}
	defer driver.Close()

	ctx := context.Background()
	memDriver := driver.(*Driver)

	// Add 3 items (at limit)
	driver.Put(ctx, "key1", "value1", 0)
	driver.Put(ctx, "key2", "value2", 0)
	driver.Put(ctx, "key3", "value3", 0)

	stats := memDriver.Stats()
	if stats.ItemCount != 3 {
		t.Errorf("Expected 3 items, got %d", stats.ItemCount)
	}

	// Add 4th item - should evict least recently used (key1)
	driver.Put(ctx, "key4", "value4", 0)

	stats = memDriver.Stats()
	if stats.ItemCount != 3 {
		t.Errorf("Expected 3 items after eviction, got %d", stats.ItemCount)
	}
	if stats.Evictions != 1 {
		t.Errorf("Expected 1 eviction, got %d", stats.Evictions)
	}

	// key1 should be evicted
	_, err = driver.Get(ctx, "key1")
	if err != dgcache.ErrKeyNotFound {
		t.Error("key1 should have been evicted")
	}

	// Other keys should still exist
	if _, err := driver.Get(ctx, "key2"); err != nil {
		t.Error("key2 should still exist")
	}
}

func TestDriver_LRUEviction(t *testing.T) {
	config := dgcache.StoreConfig{
		Driver: "memory",
		Options: map[string]interface{}{
			"max_items":       3,
			"eviction_policy": "lru",
			"enable_metrics":  true,
		},
	}

	driver, err := NewDriver(config)
	if err != nil {
		t.Fatalf("Failed to create driver: %v", err)
	}
	defer driver.Close()

	ctx := context.Background()

	// Add 3 items
	driver.Put(ctx, "key1", "value1", 0)
	driver.Put(ctx, "key2", "value2", 0)
	driver.Put(ctx, "key3", "value3", 0)

	// Access key1 (moves it to front)
	driver.Get(ctx, "key1")

	// Add key4 - should evict key2 (least recently used)
	driver.Put(ctx, "key4", "value4", 0)

	// key2 should be evicted
	_, err = driver.Get(ctx, "key2")
	if err != dgcache.ErrKeyNotFound {
		t.Error("key2 should have been evicted (LRU)")
	}

	// key1 should still exist (was accessed recently)
	if _, err := driver.Get(ctx, "key1"); err != nil {
		t.Error("key1 should still exist")
	}
}

func TestDriver_MaxBytesEviction(t *testing.T) {
	config := dgcache.StoreConfig{
		Driver: "memory",
		Options: map[string]interface{}{
			"max_bytes":      100, // 100 byte limit
			"enable_metrics": true,
		},
	}

	driver, err := NewDriver(config)
	if err != nil {
		t.Fatalf("Failed to create driver: %v", err)
	}
	defer driver.Close()

	ctx := context.Background()
	memDriver := driver.(*Driver)

	// Add items totaling 90 bytes
	driver.Put(ctx, "key1", "1234567890123456789012345678901234567890", 0) // 40 bytes
	driver.Put(ctx, "key2", "1234567890123456789012345678901234567890", 0) // 40 bytes
	driver.Put(ctx, "key3", "1234567890", 0)                               // 10 bytes
	// Total: 90 bytes

	stats := memDriver.Stats()
	if stats.ItemCount != 3 {
		t.Errorf("Expected 3 items, got %d", stats.ItemCount)
	}

	// Add item that would exceed limit - should evict
	driver.Put(ctx, "key4", "12345678901234567890123456789012345678901234567890", 0) // 50 bytes
	// Would be 140 bytes total, should evict to stay under 100

	stats = memDriver.Stats()
	if stats.Evictions == 0 {
		t.Errorf("Expected evictions due to byte limit. BytesUsed: %d, MaxBytes: 100", stats.BytesUsed)
	}
	if stats.BytesUsed > 100 {
		t.Errorf("Expected bytes used <= 100, got %d", stats.BytesUsed)
	}
}

func TestDriver_Metrics(t *testing.T) {
	config := dgcache.StoreConfig{
		Driver: "memory",
		Options: map[string]interface{}{
			"enable_metrics": true,
		},
	}

	driver, err := NewDriver(config)
	if err != nil {
		t.Fatalf("Failed to create driver: %v", err)
	}
	defer driver.Close()

	ctx := context.Background()
	memDriver := driver.(*Driver)

	// Perform operations
	driver.Put(ctx, "key1", "value1", 0)
	driver.Put(ctx, "key2", "value2", 0)
	driver.Get(ctx, "key1") // hit
	driver.Get(ctx, "key3") // miss
	driver.Forget(ctx, "key1")

	stats := memDriver.Stats()

	if stats.Sets != 2 {
		t.Errorf("Expected 2 sets, got %d", stats.Sets)
	}
	if stats.Hits != 1 {
		t.Errorf("Expected 1 hit, got %d", stats.Hits)
	}
	if stats.Misses != 1 {
		t.Errorf("Expected 1 miss, got %d", stats.Misses)
	}
	if stats.HitRate != 0.5 {
		t.Errorf("Expected 50%% hit rate, got %.2f", stats.HitRate)
	}
}

func TestDriver_NoMetrics(t *testing.T) {
	config := dgcache.StoreConfig{
		Driver: "memory",
		Options: map[string]interface{}{
			"enable_metrics": false,
		},
	}

	driver, err := NewDriver(config)
	if err != nil {
		t.Fatalf("Failed to create driver: %v", err)
	}
	defer driver.Close()

	memDriver := driver.(*Driver)

	// Stats should return empty when metrics disabled
	stats := memDriver.Stats()
	if stats.Hits != 0 || stats.Misses != 0 || stats.Sets != 0 {
		t.Error("Stats should be empty when metrics disabled")
	}
}

func TestDriver_ConfigurableCleanup(t *testing.T) {
	config := dgcache.StoreConfig{
		Driver: "memory",
		Options: map[string]interface{}{
			"cleanup_interval": 100 * time.Millisecond,
		},
	}

	driver, err := NewDriver(config)
	if err != nil {
		t.Fatalf("Failed to create driver: %v", err)
	}
	defer driver.Close()

	ctx := context.Background()

	// Add item with short TTL
	driver.Put(ctx, "key1", "value1", 50*time.Millisecond)

	// Should exist immediately
	if _, err := driver.Get(ctx, "key1"); err != nil {
		t.Error("key1 should exist")
	}

	// Wait for expiration + cleanup
	time.Sleep(200 * time.Millisecond)

	// Should be cleaned up
	_, err = driver.Get(ctx, "key1")
	if err != dgcache.ErrKeyNotFound {
		t.Error("key1 should have been cleaned up")
	}
}

func TestDriver_UpdateExisting(t *testing.T) {
	config := dgcache.StoreConfig{
		Driver: "memory",
		Options: map[string]interface{}{
			"enable_metrics": true,
		},
	}

	driver, err := NewDriver(config)
	if err != nil {
		t.Fatalf("Failed to create driver: %v", err)
	}
	defer driver.Close()

	ctx := context.Background()
	memDriver := driver.(*Driver)

	// Add item
	driver.Put(ctx, "key1", "short", 0)

	stats := memDriver.Stats()
	initialBytes := stats.BytesUsed

	// Update with larger value
	driver.Put(ctx, "key1", "much longer value", 0)

	stats = memDriver.Stats()
	if stats.ItemCount != 1 {
		t.Errorf("Expected 1 item, got %d", stats.ItemCount)
	}
	if stats.BytesUsed <= initialBytes {
		t.Error("Bytes used should have increased")
	}
}

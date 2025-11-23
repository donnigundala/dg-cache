package memory

import (
	"context"
	"sync"
	"time"

	cache "github.com/donnigundala/dg-cache"
)

// Driver is an in-memory cache driver.
// It stores cache items in memory with TTL support.
// This driver is primarily intended for testing and development.
type Driver struct {
	items  map[string]*cache.Item
	mu     sync.RWMutex
	prefix string
	ticker *time.Ticker
	done   chan bool
}

// NewDriver creates a new in-memory cache driver.
func NewDriver(config cache.StoreConfig) (cache.Driver, error) {
	d := &Driver{
		items:  make(map[string]*cache.Item),
		prefix: "",
		done:   make(chan bool),
	}

	// Start cleanup goroutine
	d.ticker = time.NewTicker(1 * time.Minute)
	go d.cleanup()

	return d, nil
}

// cleanup removes expired items periodically.
func (d *Driver) cleanup() {
	for {
		select {
		case <-d.ticker.C:
			d.removeExpired()
		case <-d.done:
			return
		}
	}
}

// removeExpired removes all expired items from the cache.
func (d *Driver) removeExpired() {
	d.mu.Lock()
	defer d.mu.Unlock()

	now := time.Now()
	for key, item := range d.items {
		if !item.ExpiresAt.IsZero() && item.ExpiresAt.Before(now) {
			delete(d.items, key)
		}
	}
}

// prefixKey adds the prefix to the key.
func (d *Driver) prefixKey(key string) string {
	if d.prefix == "" {
		return key
	}
	return d.prefix + ":" + key
}

// Get retrieves a value from the cache.
func (d *Driver) Get(ctx context.Context, key string) (interface{}, error) {
	d.mu.RLock()
	defer d.mu.RUnlock()

	item, ok := d.items[d.prefixKey(key)]
	if !ok {
		return nil, cache.ErrKeyNotFound
	}

	if item.IsExpired() {
		return nil, cache.ErrKeyNotFound
	}

	return item.Value, nil
}

// GetMultiple retrieves multiple values from the cache.
func (d *Driver) GetMultiple(ctx context.Context, keys []string) (map[string]interface{}, error) {
	d.mu.RLock()
	defer d.mu.RUnlock()

	result := make(map[string]interface{})
	for _, key := range keys {
		item, ok := d.items[d.prefixKey(key)]
		if ok && !item.IsExpired() {
			result[key] = item.Value
		}
	}

	return result, nil
}

// Put stores a value in the cache with the given TTL.
func (d *Driver) Put(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	item := &cache.Item{
		Key:   key,
		Value: value,
	}

	if ttl > 0 {
		item.ExpiresAt = time.Now().Add(ttl)
	}

	d.items[d.prefixKey(key)] = item
	return nil
}

// PutMultiple stores multiple values in the cache.
func (d *Driver) PutMultiple(ctx context.Context, items map[string]interface{}, ttl time.Duration) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	expiresAt := time.Time{}
	if ttl > 0 {
		expiresAt = time.Now().Add(ttl)
	}

	for key, value := range items {
		item := &cache.Item{
			Key:       key,
			Value:     value,
			ExpiresAt: expiresAt,
		}
		d.items[d.prefixKey(key)] = item
	}

	return nil
}

// Increment increments the value of a key.
func (d *Driver) Increment(ctx context.Context, key string, value int64) (int64, error) {
	d.mu.Lock()
	defer d.mu.Unlock()

	prefixedKey := d.prefixKey(key)
	item, ok := d.items[prefixedKey]

	var current int64
	if ok && !item.IsExpired() {
		if v, ok := item.Value.(int64); ok {
			current = v
		}
	}

	newValue := current + value
	d.items[prefixedKey] = &cache.Item{
		Key:   key,
		Value: newValue,
	}

	return newValue, nil
}

// Decrement decrements the value of a key.
func (d *Driver) Decrement(ctx context.Context, key string, value int64) (int64, error) {
	return d.Increment(ctx, key, -value)
}

// Forever stores a value in the cache indefinitely.
func (d *Driver) Forever(ctx context.Context, key string, value interface{}) error {
	return d.Put(ctx, key, value, 0)
}

// Forget removes a value from the cache.
func (d *Driver) Forget(ctx context.Context, key string) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	delete(d.items, d.prefixKey(key))
	return nil
}

// Flush removes all items from the cache.
func (d *Driver) Flush(ctx context.Context) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	d.items = make(map[string]*cache.Item)
	return nil
}

// Has checks if a key exists in the cache.
func (d *Driver) Has(ctx context.Context, key string) (bool, error) {
	d.mu.RLock()
	defer d.mu.RUnlock()

	item, ok := d.items[d.prefixKey(key)]
	if !ok {
		return false, nil
	}

	return !item.IsExpired(), nil
}

// Missing checks if a key does not exist in the cache.
func (d *Driver) Missing(ctx context.Context, key string) (bool, error) {
	has, err := d.Has(ctx, key)
	return !has, err
}

// GetPrefix returns the cache key prefix.
func (d *Driver) GetPrefix() string {
	return d.prefix
}

// SetPrefix sets the cache key prefix.
func (d *Driver) SetPrefix(prefix string) {
	d.prefix = prefix
}

// Name returns the driver name.
func (d *Driver) Name() string {
	return "memory"
}

// Close closes the driver and releases resources.
func (d *Driver) Close() error {
	d.ticker.Stop()
	d.done <- true
	close(d.done)
	return nil
}

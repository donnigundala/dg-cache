package memory

import (
	"context"
	"sync"
	"time"

	cache "github.com/donnigundala/dg-cache"
)

// Driver is an in-memory cache driver.
// It stores cache items in memory with TTL support, size limits, and LRU eviction.
type Driver struct {
	items   map[string]*cache.Item
	lru     *lruList
	nodes   map[string]*lruNode            // key -> LRU node mapping
	tags    map[string]map[string]struct{} // tag -> set of keys
	keyTags map[string][]string            // key -> list of tags
	mu      sync.RWMutex
	prefix  string
	ticker  *time.Ticker
	done    chan bool

	config  Config
	metrics *Metrics
}

// NewDriver creates a new in-memory cache driver.
func NewDriver(storeConfig cache.StoreConfig) (cache.Driver, error) {
	config := DefaultConfig()

	// Parse options from storeConfig
	if val, ok := storeConfig.Options["max_items"].(int); ok {
		config.MaxItems = val
	}
	// Handle both int and int64 for max_bytes
	if val, ok := storeConfig.Options["max_bytes"].(int64); ok {
		config.MaxBytes = val
	} else if val, ok := storeConfig.Options["max_bytes"].(int); ok {
		config.MaxBytes = int64(val)
	}
	if val, ok := storeConfig.Options["eviction_policy"].(string); ok {
		config.EvictionPolicy = val
	}
	if val, ok := storeConfig.Options["cleanup_interval"]; ok {
		if duration, ok := val.(time.Duration); ok {
			config.CleanupInterval = duration
		}
	}
	if val, ok := storeConfig.Options["enable_metrics"].(bool); ok {
		config.EnableMetrics = val
	}

	d := &Driver{
		items:   make(map[string]*cache.Item),
		lru:     newLRUList(),
		nodes:   make(map[string]*lruNode),
		tags:    make(map[string]map[string]struct{}),
		keyTags: make(map[string][]string),
		prefix:  "",
		done:    make(chan bool),
		config:  config,
	}

	if config.EnableMetrics {
		d.metrics = newMetrics()
	}

	// Start cleanup goroutine
	d.ticker = time.NewTicker(config.CleanupInterval)
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
			d.removeKeyTags(key)
			delete(d.items, key)
			delete(d.nodes, key)
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

// estimateSize estimates the size of a value in bytes.
func (d *Driver) estimateSize(value interface{}) int64 {
	switch v := value.(type) {
	case string:
		return int64(len(v))
	case []byte:
		return int64(len(v))
	case int, int8, int16, int32, int64:
		return 8
	case uint, uint8, uint16, uint32, uint64:
		return 8
	case float32, float64:
		return 8
	case bool:
		return 1
	default:
		// Default estimate for complex types
		return 64
	}
}

// evictIfNeeded evicts items if size limits would be exceeded by adding newItemSize bytes.
func (d *Driver) evictIfNeeded(newItemSize int64) {
	// Check item count limit
	if d.config.MaxItems > 0 && len(d.items) >= d.config.MaxItems {
		d.evictOne()
	}

	// Check bytes limit - evict until we have room for the new item
	if d.config.MaxBytes > 0 {
		// Calculate current size
		currentBytes := int64(0)
		if d.metrics != nil {
			currentBytes = d.metrics.bytesUsed
		} else {
			// Calculate on the fly if metrics disabled
			for _, item := range d.items {
				currentBytes += d.estimateSize(item.Value)
			}
		}

		for currentBytes+newItemSize > d.config.MaxBytes {
			if !d.evictOne() {
				break // No more items to evict
			}
			// Recalculate current size after eviction
			if d.metrics != nil {
				currentBytes = d.metrics.bytesUsed
			} else {
				currentBytes = 0
				for _, item := range d.items {
					currentBytes += d.estimateSize(item.Value)
				}
			}
		}
	}
}

// evictOne evicts a single item based on the eviction policy.
// Returns true if an item was evicted, false if cache is empty.
func (d *Driver) evictOne() bool {
	if d.config.EvictionPolicy == "lru" {
		key := d.lru.removeLast()
		if key == "" {
			return false
		}

		if item, ok := d.items[key]; ok {
			size := d.estimateSize(item.Value)
			if d.metrics != nil {
				d.metrics.RecordEviction(size)
			}
			d.removeKeyTags(key)
			delete(d.items, key)
			delete(d.nodes, key)
			return true
		}
	}
	return false
}

// Get retrieves a value from the cache.
func (d *Driver) Get(ctx context.Context, key string) (interface{}, error) {
	d.mu.Lock()
	defer d.mu.Unlock()

	prefixedKey := d.prefixKey(key)
	item, ok := d.items[prefixedKey]

	if !ok || item.IsExpired() {
		if d.metrics != nil {
			d.metrics.RecordMiss()
		}
		return nil, cache.ErrKeyNotFound
	}

	// Update LRU
	if node, ok := d.nodes[prefixedKey]; ok {
		d.lru.moveToFront(node)
	}

	if d.metrics != nil {
		d.metrics.RecordHit()
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
	return d.put(key, value, ttl)
}

// put is the internal unlocked implementation of Put.
func (d *Driver) put(key string, value interface{}, ttl time.Duration) error {
	prefixedKey := d.prefixKey(key)
	newSize := d.estimateSize(value)

	// Calculate net size change (for replacements)
	netSizeChange := newSize
	if oldItem, ok := d.items[prefixedKey]; ok {
		oldSize := d.estimateSize(oldItem.Value)
		netSizeChange = newSize - oldSize
	}

	// Check if we need to evict (pass the net size change)
	if netSizeChange > 0 {
		d.evictIfNeeded(netSizeChange)
	}

	item := &cache.Item{
		Key:   key,
		Value: value,
	}

	if ttl > 0 {
		item.ExpiresAt = time.Now().Add(ttl)
	}

	// Update metrics
	if d.metrics != nil {
		if oldItem, ok := d.items[prefixedKey]; ok {
			// Replacing existing item
			oldSize := d.estimateSize(oldItem.Value)
			d.metrics.RecordUpdate(oldSize, newSize)
		} else {
			d.metrics.RecordSet(newSize)
		}
	}

	d.items[prefixedKey] = item

	// Update LRU
	if node, ok := d.nodes[prefixedKey]; ok {
		d.lru.moveToFront(node)
	} else {
		d.nodes[prefixedKey] = d.lru.addToFront(prefixedKey)
	}

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
	return d.forget(key)
}

// forget is the internal unlocked implementation of Forget.
func (d *Driver) forget(key string) error {
	prefixedKey := d.prefixKey(key)
	d.removeKeyTags(prefixedKey)
	delete(d.items, prefixedKey)
	delete(d.nodes, prefixedKey)
	return nil
}

// ForgetMultiple removes multiple values from the cache.
func (d *Driver) ForgetMultiple(ctx context.Context, keys []string) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	for _, key := range keys {
		prefixedKey := d.prefixKey(key)
		d.removeKeyTags(prefixedKey)
		delete(d.items, prefixedKey)
		delete(d.nodes, prefixedKey)
	}
	return nil
}

// Flush removes all items from the cache.
func (d *Driver) Flush(ctx context.Context) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	// Clear everything
	d.items = make(map[string]*cache.Item)
	d.nodes = make(map[string]*lruNode)
	d.lru = newLRUList()
	d.tags = make(map[string]map[string]struct{})
	d.keyTags = make(map[string][]string)
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

// Stats returns a snapshot of current cache statistics.
func (d *Driver) Stats() cache.Stats {
	if d.metrics == nil {
		return cache.Stats{}
	}
	return d.metrics.Stats()
}

// Close closes the driver and releases resources.
func (d *Driver) Close() error {
	d.ticker.Stop()
	d.done <- true
	close(d.done)
	return nil
}

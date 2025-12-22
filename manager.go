package dgcache

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/donnigundala/dg-core/contracts/cache"
	"go.opentelemetry.io/otel/metric"
)

// Manager manages multiple cache stores and provides a unified interface.
type Manager struct {
	config       Config
	stores       map[string]cache.Store
	drivers      map[string]DriverFactory
	mu           sync.RWMutex
	defaultStore string

	// Observability
	metricHits      metric.Int64ObservableCounter
	metricMisses    metric.Int64ObservableCounter
	metricSets      metric.Int64ObservableCounter
	metricDeletes   metric.Int64ObservableCounter
	metricEvictions metric.Int64ObservableCounter
	metricItems     metric.Int64ObservableGauge
	metricBytes     metric.Int64ObservableGauge
}

// DriverFactory is a function that creates a cache driver.
type DriverFactory func(config StoreConfig) (cache.Driver, error)

// NewManager creates a new cache manager with the given configuration.
func NewManager(config Config) (*Manager, error) {
	if err := config.Validate(); err != nil {
		return nil, err
	}

	m := &Manager{
		config:       config,
		stores:       make(map[string]cache.Store),
		drivers:      make(map[string]DriverFactory),
		defaultStore: config.DefaultStore,
	}

	return m, nil
}

// RegisterDriver registers a driver factory for the given driver name.
func (m *Manager) RegisterDriver(name string, factory DriverFactory) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.drivers[name] = factory
}

// Verify Manager implements Cache interface
var _ cache.Cache = (*Manager)(nil)

// DefaultStore returns the default cache store.
func (m *Manager) DefaultStore() cache.Store {
	store, _ := m.Store("")
	return store
}

// Store returns the cache store with the given name.
// If name is empty, returns the default store.
func (m *Manager) Store(name string) (cache.Store, error) {
	if name == "" {
		name = m.defaultStore
	}

	m.mu.RLock()
	store, ok := m.stores[name]
	m.mu.RUnlock()

	if ok {
		return store, nil
	}

	// Store not initialized, create it
	return m.createStore(name)
}

// createStore creates and caches a new store instance.
func (m *Manager) createStore(name string) (cache.Store, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Double-check after acquiring write lock
	if store, ok := m.stores[name]; ok {
		return store, nil
	}

	// Get store config
	storeConfig, ok := m.config.Stores[name]
	if !ok {
		return nil, ErrStoreNotFound
	}

	// Get driver factory
	factory, ok := m.drivers[storeConfig.Driver]
	if !ok {
		return nil, ErrDriverNotFound
	}

	// Create driver
	driver, err := factory(storeConfig)
	if err != nil {
		return nil, ErrDriverError(storeConfig.Driver, err)
	}

	// Set prefix
	prefix := storeConfig.Prefix
	if prefix == "" {
		prefix = m.config.Prefix
	}
	driver.SetPrefix(prefix)

	// Cache the store
	m.stores[name] = driver

	return driver, nil
}

// Get retrieves a value from the default cache store.
func (m *Manager) Get(ctx context.Context, key string) (interface{}, error) {
	store, err := m.Store("")
	if err != nil {
		return nil, err
	}
	return store.Get(ctx, key)
}

// GetMultiple retrieves multiple values from the default cache store.
func (m *Manager) GetMultiple(ctx context.Context, keys []string) (map[string]interface{}, error) {
	store, err := m.Store("")
	if err != nil {
		return nil, err
	}
	return store.GetMultiple(ctx, keys)
}

// Put stores a value in the default cache store.
func (m *Manager) Put(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	store, err := m.Store("")
	if err != nil {
		return err
	}
	return store.Put(ctx, key, value, ttl)
}

// PutMultiple stores multiple values in the default cache store.
func (m *Manager) PutMultiple(ctx context.Context, items map[string]interface{}, ttl time.Duration) error {
	store, err := m.Store("")
	if err != nil {
		return err
	}
	return store.PutMultiple(ctx, items, ttl)
}

// Increment increments a value in the default cache store.
func (m *Manager) Increment(ctx context.Context, key string, value int64) (int64, error) {
	store, err := m.Store("")
	if err != nil {
		return 0, err
	}
	return store.Increment(ctx, key, value)
}

// Decrement decrements a value in the default cache store.
func (m *Manager) Decrement(ctx context.Context, key string, value int64) (int64, error) {
	store, err := m.Store("")
	if err != nil {
		return 0, err
	}
	return store.Decrement(ctx, key, value)
}

// Forever stores a value in the default cache store indefinitely.
func (m *Manager) Forever(ctx context.Context, key string, value interface{}) error {
	store, err := m.Store("")
	if err != nil {
		return err
	}
	return store.Forever(ctx, key, value)
}

// Forget removes a value from the default cache store.
func (m *Manager) Forget(ctx context.Context, key string) error {
	store, err := m.Store("")
	if err != nil {
		return err
	}
	return store.Forget(ctx, key)
}

// ForgetMultiple removes multiple values from the default cache store.
func (m *Manager) ForgetMultiple(ctx context.Context, keys []string) error {
	store, err := m.Store("")
	if err != nil {
		return err
	}
	return store.ForgetMultiple(ctx, keys)
}

// Flush removes all items from the default cache store.
func (m *Manager) Flush(ctx context.Context) error {
	store, err := m.Store("")
	if err != nil {
		return err
	}
	return store.Flush(ctx)
}

// Has checks if a key exists in the default cache store.
func (m *Manager) Has(ctx context.Context, key string) (bool, error) {
	store, err := m.Store("")
	if err != nil {
		return false, err
	}
	return store.Has(ctx, key)
}

// Stats returns the statistics of the default cache store.
func (m *Manager) Stats() cache.Stats {
	store, err := m.Store("")
	if err != nil {
		return cache.Stats{}
	}
	return store.Stats()
}

// Tags returns a tagged cache store.
func (m *Manager) Tags(tags ...string) cache.TaggedStore {
	store, err := m.Store("")
	if err != nil {
		panic(fmt.Sprintf("failed to get default store: %v", err))
	}
	if Taggable, ok := store.(cache.TaggedStore); ok {
		return Taggable.Tags(tags...)
	}
	panic("default cache store does not support tagging")
}

// Missing checks if a key does not exist in the default cache store.
func (m *Manager) Missing(ctx context.Context, key string) (bool, error) {
	store, err := m.Store("")
	if err != nil {
		return false, err
	}
	return store.Missing(ctx, key)
}

// Remember retrieves a value from the cache or executes the callback and stores the result.
// This implements the cache-aside pattern.
func (m *Manager) Remember(ctx context.Context, key string, ttl time.Duration, callback func() (interface{}, error)) (interface{}, error) {
	// Try to get from cache
	value, err := m.Get(ctx, key)
	if err == nil && value != nil {
		return value, nil
	}

	// Execute callback
	value, err = callback()
	if err != nil {
		return nil, err
	}

	// Store in cache
	if err := m.Put(ctx, key, value, ttl); err != nil {
		// Log error but don't fail - we have the value
		return value, nil
	}

	return value, nil
}

// RememberForever retrieves a value from the cache or executes the callback and stores the result forever.
func (m *Manager) RememberForever(ctx context.Context, key string, callback func() (interface{}, error)) (interface{}, error) {
	// Try to get from cache
	value, err := m.Get(ctx, key)
	if err == nil && value != nil {
		return value, nil
	}

	// Execute callback
	value, err = callback()
	if err != nil {
		return nil, err
	}

	// Store in cache forever
	if err := m.Forever(ctx, key, value); err != nil {
		// Log error but don't fail - we have the value
		return value, nil
	}

	return value, nil
}

// Pull retrieves a value from the cache and then deletes it.
func (m *Manager) Pull(ctx context.Context, key string) (interface{}, error) {
	value, err := m.Get(ctx, key)
	if err != nil {
		return nil, err
	}

	// Delete the key (ignore errors)
	_ = m.Forget(ctx, key)

	return value, nil
}

// GetPrefix returns the prefix of the default store.
func (m *Manager) GetPrefix() string {
	return m.DefaultStore().GetPrefix()
}

// SetPrefix sets the prefix of the default store.
func (m *Manager) SetPrefix(prefix string) {
	m.DefaultStore().SetPrefix(prefix)
}

// Stop stops the cache manager gracefully.
// This implements the Stoppable interface.
func (m *Manager) Stop(ctx context.Context) error {
	return m.Close()
}

// Close closes all cache stores and releases resources.
func (m *Manager) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	var lastErr error
	for name, store := range m.stores {
		if driver, ok := store.(cache.Driver); ok {
			if err := driver.Close(); err != nil {
				lastErr = err
			}
		}
		delete(m.stores, name)
	}

	return lastErr
}

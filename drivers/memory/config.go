package memory

import "time"

// Config represents the configuration for the memory cache driver.
type Config struct {
	// MaxItems is the maximum number of items in the cache.
	// 0 means unlimited (default).
	MaxItems int

	// MaxBytes is the maximum total size of cached items in bytes.
	// 0 means unlimited (default).
	MaxBytes int64

	// EvictionPolicy determines how items are evicted when limits are reached.
	// Options: "lru" (default), "lfu", "fifo"
	EvictionPolicy string

	// CleanupInterval is how often expired items are cleaned up.
	// Default: 1 minute
	CleanupInterval time.Duration

	// EnableMetrics enables collection of cache statistics.
	// Default: false
	EnableMetrics bool
}

// DefaultConfig returns a default memory cache configuration.
func DefaultConfig() Config {
	return Config{
		MaxItems:        0, // unlimited
		MaxBytes:        0, // unlimited
		EvictionPolicy:  "lru",
		CleanupInterval: 1 * time.Minute,
		EnableMetrics:   false,
	}
}

// WithMaxItems sets the maximum number of items.
func (c Config) WithMaxItems(max int) Config {
	c.MaxItems = max
	return c
}

// WithMaxBytes sets the maximum total size in bytes.
func (c Config) WithMaxBytes(max int64) Config {
	c.MaxBytes = max
	return c
}

// WithEvictionPolicy sets the eviction policy.
func (c Config) WithEvictionPolicy(policy string) Config {
	c.EvictionPolicy = policy
	return c
}

// WithCleanupInterval sets the cleanup interval.
func (c Config) WithCleanupInterval(interval time.Duration) Config {
	c.CleanupInterval = interval
	return c
}

// WithMetrics enables or disables metrics collection.
func (c Config) WithMetrics(enabled bool) Config {
	c.EnableMetrics = enabled
	return c
}

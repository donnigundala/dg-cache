package cache

import (
	"context"
	"time"
)

// Store defines the interface for cache operations.
// All cache drivers must implement this interface.
type Store interface {
	// Get retrieves a value from the cache by key.
	// Returns nil if the key doesn't exist or has expired.
	Get(ctx context.Context, key string) (interface{}, error)

	// GetMultiple retrieves multiple values from the cache.
	// Returns a map of key-value pairs. Missing keys are not included in the result.
	GetMultiple(ctx context.Context, keys []string) (map[string]interface{}, error)

	// Put stores a value in the cache with the given TTL.
	// If ttl is 0, the item never expires (same as Forever).
	Put(ctx context.Context, key string, value interface{}, ttl time.Duration) error

	// PutMultiple stores multiple values in the cache with the same TTL.
	PutMultiple(ctx context.Context, items map[string]interface{}, ttl time.Duration) error

	// Increment increments the value of a key by the given amount.
	// Returns the new value after incrementing.
	// If the key doesn't exist, it's created with the increment value.
	Increment(ctx context.Context, key string, value int64) (int64, error)

	// Decrement decrements the value of a key by the given amount.
	// Returns the new value after decrementing.
	// If the key doesn't exist, it's created with the negative of the decrement value.
	Decrement(ctx context.Context, key string, value int64) (int64, error)

	// Forever stores a value in the cache indefinitely (no expiration).
	Forever(ctx context.Context, key string, value interface{}) error

	// Forget removes a value from the cache.
	// Returns nil if the key doesn't exist.
	Forget(ctx context.Context, key string) error

	// Flush removes all items from the cache.
	Flush(ctx context.Context) error

	// Has checks if a key exists in the cache.
	Has(ctx context.Context, key string) (bool, error)

	// Missing checks if a key does not exist in the cache.
	// This is the inverse of Has.
	Missing(ctx context.Context, key string) (bool, error)

	// GetPrefix returns the cache key prefix.
	GetPrefix() string

	// SetPrefix sets the cache key prefix.
	SetPrefix(prefix string)
}

// TaggedStore extends Store with tagging capabilities.
// Tagged caches allow grouping related cache items and flushing them together.
type TaggedStore interface {
	Store

	// Tags returns a new TaggedStore instance with the given tags.
	// Multiple calls to Tags are cumulative.
	Tags(tags ...string) TaggedStore

	// FlushTags removes all items associated with the given tags.
	FlushTags(ctx context.Context, tags ...string) error
}

// Driver is the interface that cache drivers must implement.
// It extends Store with driver-specific functionality.
type Driver interface {
	Store

	// Name returns the driver name (e.g., "redis", "memory").
	Name() string

	// Close closes the driver and releases any resources.
	Close() error
}

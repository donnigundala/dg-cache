package cache

import (
	"context"
	"time"
)

// Cache defines the interface for cache operations.
// This allows for better abstraction and testability.
// The Manager type implements this interface.
type Cache interface {
	// Core Operations
	Get(ctx context.Context, key string) (interface{}, error)
	Put(ctx context.Context, key string, value interface{}, ttl time.Duration) error
	Has(ctx context.Context, key string) (bool, error)
	Forget(ctx context.Context, key string) error
	Flush(ctx context.Context) error

	// Batch Operations
	GetMultiple(ctx context.Context, keys []string) (map[string]interface{}, error)
	PutMultiple(ctx context.Context, items map[string]interface{}, ttl time.Duration) error
	ForgetMultiple(ctx context.Context, keys []string) error

	// Advanced Operations
	Forever(ctx context.Context, key string, value interface{}) error
	Remember(ctx context.Context, key string, ttl time.Duration, callback func() (interface{}, error)) (interface{}, error)
	RememberForever(ctx context.Context, key string, callback func() (interface{}, error)) (interface{}, error)
	Pull(ctx context.Context, key string) (interface{}, error)
	Increment(ctx context.Context, key string, value int64) (int64, error)
	Decrement(ctx context.Context, key string, value int64) (int64, error)

	// Store Management
	Store(name string) (Store, error)
	DefaultStore() Store

	// Typed Helpers
	GetAs(ctx context.Context, key string, dest interface{}) error
	GetString(ctx context.Context, key string) (string, error)
	GetInt(ctx context.Context, key string) (int, error)
	GetInt64(ctx context.Context, key string) (int64, error)
	GetFloat64(ctx context.Context, key string) (float64, error)
	GetBool(ctx context.Context, key string) (bool, error)
}

// Stats represents cache statistics at a point in time.
type Stats struct {
	// Hits is the number of cache hits.
	Hits int64

	// Misses is the number of cache misses.
	Misses int64

	// Sets is the number of set operations.
	Sets int64

	// Deletes is the number of delete operations.
	Deletes int64

	// Evictions is the number of evicted items.
	Evictions int64

	// ItemCount is the current number of items in the cache.
	ItemCount int

	// BytesUsed is the estimated total size of cached items in bytes.
	BytesUsed int64

	// HitRate is the cache hit rate (hits / (hits + misses)).
	HitRate float64
}

// Observable is an interface for drivers that expose statistics.
type Observable interface {
	Stats() Stats
}

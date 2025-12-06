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
}

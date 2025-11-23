package cache

import (
	"fmt"
)

// Error types for cache operations.
var (
	// ErrKeyNotFound is returned when a cache key is not found.
	ErrKeyNotFound = fmt.Errorf("cache: key not found")

	// ErrInvalidValue is returned when a cache value is invalid.
	ErrInvalidValue = fmt.Errorf("cache: invalid value")

	// ErrDriverNotFound is returned when a cache driver is not found.
	ErrDriverNotFound = fmt.Errorf("cache: driver not found")

	// ErrStoreNotFound is returned when a cache store is not found.
	ErrStoreNotFound = fmt.Errorf("cache: store not found")
)

// ErrInvalidConfig returns a configuration error with a formatted message.
func ErrInvalidConfig(format string, args ...interface{}) error {
	return fmt.Errorf("cache: invalid config: "+format, args...)
}

// ErrDriverError returns a driver error with a formatted message.
func ErrDriverError(driver string, err error) error {
	return fmt.Errorf("cache: driver '%s' error: %w", driver, err)
}

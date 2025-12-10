package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/donnigundala/dg-core/contracts/foundation"
)

// GetAs retrieves a value and unmarshals it into the provided destination pointer.
// This provides type-safe retrieval with automatic deserialization.
//
// Example:
//
//	var user User
//	if err := cache.GetAs(ctx, "user:1", &user); err != nil {
//	    // handle error
//	}
func (m *Manager) GetAs(ctx context.Context, key string, dest interface{}) error {
	value, err := m.Get(ctx, key)
	if err != nil {
		return err
	}

	// If value is nil, return error
	if value == nil {
		return ErrKeyNotFound
	}

	// Get the type of dest
	destType := reflect.TypeOf(dest)
	if destType.Kind() != reflect.Ptr {
		return fmt.Errorf("dest must be a pointer")
	}

	// Get the type of value
	valueType := reflect.TypeOf(value)

	// If types match exactly, assign directly
	if valueType == destType.Elem() {
		reflect.ValueOf(dest).Elem().Set(reflect.ValueOf(value))
		return nil
	}

	// If value is already the right type (for interface{} cases)
	if valueType.AssignableTo(destType.Elem()) {
		reflect.ValueOf(dest).Elem().Set(reflect.ValueOf(value))
		return nil
	}

	// Try JSON unmarshal as fallback (for serialized data)
	if data, ok := value.([]byte); ok {
		return json.Unmarshal(data, dest)
	}
	if str, ok := value.(string); ok {
		return json.Unmarshal([]byte(str), dest)
	}

	return fmt.Errorf("cannot convert %T to %T", value, dest)
}

// GetString retrieves a string value from the cache.
// Returns empty string and error if key doesn't exist or value is not a string.
func (m *Manager) GetString(ctx context.Context, key string) (string, error) {
	val, err := m.Get(ctx, key)
	if err != nil {
		return "", err
	}

	if s, ok := val.(string); ok {
		return s, nil
	}

	// Try to convert to string
	return fmt.Sprintf("%v", val), nil
}

// GetInt retrieves an int value from the cache.
// Returns 0 and error if key doesn't exist or value cannot be converted to int.
func (m *Manager) GetInt(ctx context.Context, key string) (int, error) {
	val, err := m.Get(ctx, key)
	if err != nil {
		return 0, err
	}

	// Try direct type assertion
	if i, ok := val.(int); ok {
		return i, nil
	}

	// JSON unmarshals numbers as float64
	if f, ok := val.(float64); ok {
		return int(f), nil
	}

	// Try int64
	if i64, ok := val.(int64); ok {
		return int(i64), nil
	}

	return 0, fmt.Errorf("value is not an int: got %T", val)
}

// GetInt64 retrieves an int64 value from the cache.
func (m *Manager) GetInt64(ctx context.Context, key string) (int64, error) {
	val, err := m.Get(ctx, key)
	if err != nil {
		return 0, err
	}

	if i64, ok := val.(int64); ok {
		return i64, nil
	}

	if f, ok := val.(float64); ok {
		return int64(f), nil
	}

	if i, ok := val.(int); ok {
		return int64(i), nil
	}

	return 0, fmt.Errorf("value is not an int64: got %T", val)
}

// GetFloat64 retrieves a float64 value from the cache.
func (m *Manager) GetFloat64(ctx context.Context, key string) (float64, error) {
	val, err := m.Get(ctx, key)
	if err != nil {
		return 0, err
	}

	if f, ok := val.(float64); ok {
		return f, nil
	}

	if f32, ok := val.(float32); ok {
		return float64(f32), nil
	}

	if i, ok := val.(int); ok {
		return float64(i), nil
	}

	return 0, fmt.Errorf("value is not a float64: got %T", val)
}

// GetBool retrieves a bool value from the cache.
func (m *Manager) GetBool(ctx context.Context, key string) (bool, error) {
	val, err := m.Get(ctx, key)
	if err != nil {
		return false, err
	}

	if b, ok := val.(bool); ok {
		return b, nil
	}

	return false, fmt.Errorf("value is not a bool: got %T", val)
}

// -----------------------------------------------------------------------------
// Container Integration Helpers (v1.6.0)
// -----------------------------------------------------------------------------

// Resolve resolves the main cache manager from the application container.
func Resolve(app foundation.Application) (Cache, error) {
	instance, err := app.Make("cache")
	if err != nil {
		return nil, fmt.Errorf("failed to resolve cache: %w", err)
	}

	cache, ok := instance.(Cache)
	if !ok {
		return nil, fmt.Errorf("resolved instance is not a Cache")
	}

	return cache, nil
}

// MustResolve resolves the cache manager or panics.
func MustResolve(app foundation.Application) Cache {
	cache, err := Resolve(app)
	if err != nil {
		panic(err)
	}
	return cache
}

// ResolveStore resolves a named cache store from the container.
func ResolveStore(app foundation.Application, name string) (Store, error) {
	instance, err := app.Make(fmt.Sprintf("cache.%s", name))
	if err != nil {
		return nil, fmt.Errorf("failed to resolve cache store %s: %w", name, err)
	}

	store, ok := instance.(Store)
	if !ok {
		return nil, fmt.Errorf("resolved instance is not a Store")
	}

	return store, nil
}

// MustResolveStore resolves a named store or panics.
func MustResolveStore(app foundation.Application, name string) Store {
	store, err := ResolveStore(app, name)
	if err != nil {
		panic(err)
	}
	return store
}

// Injectable provides a convenient way to inject cache dependencies.
// Include this struct in your services to easily access cache stores.
type Injectable struct {
	app foundation.Application
}

// NewInjectable creates a new Injectable instance.
func NewInjectable(app foundation.Application) *Injectable {
	return &Injectable{app: app}
}

// Cache returns the main cache manager.
// Panics if cache cannot be resolved.
func (i *Injectable) Cache() Cache {
	return MustResolve(i.app)
}

// Store returns a named cache store.
// Panics if the store cannot be resolved.
func (i *Injectable) Store(name string) Store {
	return MustResolveStore(i.app, name)
}

// TryStore returns a named store or nil if it doesn't exist.
// This is safe to use for optional cache stores.
func (i *Injectable) TryStore(name string) Store {
	store, err := ResolveStore(i.app, name)
	if err != nil {
		return nil
	}
	return store
}

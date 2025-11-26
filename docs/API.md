# API Reference

## Table of Contents
- [Manager](#manager)
- [Configuration](#configuration)
- [Drivers](#drivers)
- [Serialization](#serialization)
- [Typed Helpers](#typed-helpers)

## Manager

The `Manager` is the central component for cache operations.

### Constructor

#### `NewManager(config Config) (*Manager, error)`

Creates a new cache manager.

**Parameters:**
- `config` - Cache configuration

**Returns:**
- `*Manager` - Cache manager instance
- `error` - Error if initialization fails

**Example:**
```go
config := cache.DefaultConfig()
manager, err := cache.NewManager(config)
if err != nil {
    log.Fatal(err)
}
defer manager.Close()
```

### Basic Operations

#### `Get(ctx context.Context, key string) (interface{}, error)`

Retrieves a value from the cache.

**Parameters:**
- `ctx` - Context
- `key` - Cache key

**Returns:**
- `interface{}` - Cached value
- `error` - `ErrKeyNotFound` if key doesn't exist

**Example:**
```go
val, err := manager.Get(ctx, "user:1")
if err == cache.ErrKeyNotFound {
    // Key not found
}
user := val.(User)
```

#### `Put(ctx context.Context, key string, value interface{}, ttl time.Duration) error`

Stores a value in the cache.

**Parameters:**
- `ctx` - Context
- `key` - Cache key
- `value` - Value to cache (any Go type)
- `ttl` - Time to live (0 = no expiration)

**Returns:**
- `error` - Error if operation fails

**Example:**
```go
user := User{ID: 1, Name: "John"}
err := manager.Put(ctx, "user:1", user, 1*time.Hour)
```

#### `Forget(ctx context.Context, key string) error`

Removes a value from the cache.

**Parameters:**
- `ctx` - Context
- `key` - Cache key

**Returns:**
- `error` - Error if operation fails

**Example:**
```go
err := manager.Forget(ctx, "user:1")
```

#### `Flush(ctx context.Context) error`

Clears all items from the cache.

**Returns:**
- `error` - Error if operation fails

**Example:**
```go
err := manager.Flush(ctx)
```

#### `Has(ctx context.Context, key string) (bool, error)`

Checks if a key exists in the cache.

**Parameters:**
- `ctx` - Context
- `key` - Cache key

**Returns:**
- `bool` - True if key exists
- `error` - Error if operation fails

**Example:**
```go
exists, err := manager.Has(ctx, "user:1")
if exists {
    // Key exists
}
```

#### `Missing(ctx context.Context, key string) (bool, error)`

Checks if a key is missing from the cache.

**Parameters:**
- `ctx` - Context
- `key` - Cache key

**Returns:**
- `bool` - True if key doesn't exist
- `error` - Error if operation fails

**Example:**
```go
missing, err := manager.Missing(ctx, "user:1")
if missing {
    // Key doesn't exist
}
```

#### `Pull(ctx context.Context, key string) (interface{}, error)`

Retrieves and removes a value from the cache.

**Parameters:**
- `ctx` - Context
- `key` - Cache key

**Returns:**
- `interface{}` - Cached value
- `error` - Error if operation fails

**Example:**
```go
val, err := manager.Pull(ctx, "temp_token")
```

### Batch Operations

#### `GetMultiple(ctx context.Context, keys []string) (map[string]interface{}, error)`

Retrieves multiple values from the cache.

**Parameters:**
- `ctx` - Context
- `keys` - Array of cache keys

**Returns:**
- `map[string]interface{}` - Map of key-value pairs
- `error` - Error if operation fails

**Example:**
```go
keys := []string{"user:1", "user:2", "user:3"}
values, err := manager.GetMultiple(ctx, keys)
for key, val := range values {
    user := val.(User)
    fmt.Printf("%s: %+v\n", key, user)
}
```

#### `PutMultiple(ctx context.Context, items map[string]interface{}, ttl time.Duration) error`

Stores multiple values in the cache.

**Parameters:**
- `ctx` - Context
- `items` - Map of key-value pairs
- `ttl` - Time to live for all items

**Returns:**
- `error` - Error if operation fails

**Example:**
```go
items := map[string]interface{}{
    "user:1": User{ID: 1, Name: "John"},
    "user:2": User{ID: 2, Name: "Jane"},
}
err := manager.PutMultiple(ctx, items, 1*time.Hour)
```

### Atomic Operations

#### `Increment(ctx context.Context, key string, value int64) (int64, error)`

Increments a numeric value.

**Parameters:**
- `ctx` - Context
- `key` - Cache key
- `value` - Amount to increment

**Returns:**
- `int64` - New value after increment
- `error` - Error if operation fails

**Example:**
```go
newVal, err := manager.Increment(ctx, "page_views", 1)
fmt.Printf("Page views: %d\n", newVal)
```

#### `Decrement(ctx context.Context, key string, value int64) (int64, error)`

Decrements a numeric value.

**Parameters:**
- `ctx` - Context
- `key` - Cache key
- `value` - Amount to decrement

**Returns:**
- `int64` - New value after decrement
- `error` - Error if operation fails

**Example:**
```go
newVal, err := manager.Decrement(ctx, "stock_count", 1)
```

### Remember Pattern

#### `Remember(ctx context.Context, key string, ttl time.Duration, callback func() (interface{}, error)) (interface{}, error)`

Retrieves a value from cache, or executes callback and caches the result if not found.

**Parameters:**
- `ctx` - Context
- `key` - Cache key
- `ttl` - Time to live for cached result
- `callback` - Function to execute if key not found

**Returns:**
- `interface{}` - Cached or computed value
- `error` - Error if operation fails

**Example:**
```go
user, err := manager.Remember(ctx, "user:1", 1*time.Hour, func() (interface{}, error) {
    return db.FindUser(1)
})
```

#### `RememberForever(ctx context.Context, key string, callback func() (interface{}, error)) (interface{}, error)`

Like Remember, but caches the result forever (no expiration).

**Example:**
```go
config, err := manager.RememberForever(ctx, "app_config", func() (interface{}, error) {
    return loadConfig()
})
```

### Typed Helpers

#### `GetAs(ctx context.Context, key string, dest interface{}) error`

Retrieves a value and unmarshals it into the provided destination pointer.

**Parameters:**
- `ctx` - Context
- `key` - Cache key
- `dest` - Pointer to destination variable

**Returns:**
- `error` - Error if operation fails or type mismatch

**Example:**
```go
var user User
if err := manager.GetAs(ctx, "user:1", &user); err != nil {
    // Handle error
}
```

#### `GetString(ctx context.Context, key string) (string, error)`

Retrieves a string value.

**Returns:**
- `string` - String value
- `error` - Error if operation fails

**Example:**
```go
name, err := manager.GetString(ctx, "user_name")
```

#### `GetInt(ctx context.Context, key string) (int, error)`

Retrieves an int value.

**Returns:**
- `int` - Integer value
- `error` - Error if operation fails

**Example:**
```go
age, err := manager.GetInt(ctx, "user_age")
```

#### `GetInt64(ctx context.Context, key string) (int64, error)`

Retrieves an int64 value.

**Returns:**
- `int64` - Integer value
- `error` - Error if operation fails

#### `GetFloat64(ctx context.Context, key string) (float64, error)`

Retrieves a float64 value.

**Returns:**
- `float64` - Float value
- `error` - Error if operation fails

**Example:**
```go
price, err := manager.GetFloat64(ctx, "product_price")
```

#### `GetBool(ctx context.Context, key string) (bool, error)`

Retrieves a bool value.

**Returns:**
- `bool` - Boolean value
- `error` - Error if operation fails

**Example:**
```go
active, err := manager.GetBool(ctx, "user_active")
```

### Store Management

#### `Store(name string) (Driver, error)`

Returns a named cache store.

**Parameters:**
- `name` - Store name

**Returns:**
- `Driver` - Cache driver instance
- `error` - Error if store not found

**Example:**
```go
redisStore, err := manager.Store("redis")
redisStore.Put(ctx, "key", "value", 0)
```

#### `RegisterDriver(name string, factory DriverFactory)`

Registers a cache driver.

**Parameters:**
- `name` - Driver name
- `factory` - Driver factory function

**Example:**
```go
manager.RegisterDriver("memory", memory.NewDriver)
manager.RegisterDriver("redis", redis.NewDriver)
```

#### `Close() error`

Closes all cache connections.

**Returns:**
- `error` - Error if close fails

**Example:**
```go
defer manager.Close()
```

## Configuration

### Config Struct

```go
type Config struct {
    DefaultStore string
    Prefix       string
    Stores       map[string]StoreConfig
}

type StoreConfig struct {
    Driver     string
    Connection string
    Prefix     string
    Options    map[string]interface{}
}
```

### Default Configuration

#### `DefaultConfig() Config`

Returns a configuration with sensible defaults.

**Example:**
```go
config := cache.DefaultConfig()
```

### Store Configuration

#### Memory Driver Options

```go
Options: map[string]interface{}{
    "max_items":        1000,              // Maximum number of items
    "max_bytes":        10 * 1024 * 1024,  // Maximum total size (10MB)
    "eviction_policy":  "lru",             // Eviction policy
    "cleanup_interval": 1 * time.Minute,   // Cleanup interval
    "enable_metrics":   true,              // Enable metrics
}
```

#### Redis Driver Options

```go
Options: map[string]interface{}{
    "host":       "localhost",
    "port":       6379,
    "password":   "",
    "database":   0,
    "pool_size":  10,
    "serializer": "msgpack",  // or "json"
}
```

## Drivers

### Driver Interface

```go
type Driver interface {
    Get(ctx context.Context, key string) (interface{}, error)
    Put(ctx context.Context, key string, value interface{}, ttl time.Duration) error
    Forget(ctx context.Context, key string) error
    Flush(ctx context.Context) error
    Has(ctx context.Context, key string) (bool, error)
    Missing(ctx context.Context, key string) (bool, error)
    Pull(ctx context.Context, key string) (interface{}, error)
    GetMultiple(ctx context.Context, keys []string) (map[string]interface{}, error)
    PutMultiple(ctx context.Context, items map[string]interface{}, ttl time.Duration) error
    Increment(ctx context.Context, key string, value int64) (int64, error)
    Decrement(ctx context.Context, key string, value int64) (int64, error)
    Forever(ctx context.Context, key string, value interface{}) error
    GetPrefix() string
    SetPrefix(prefix string)
    Name() string
    Close() error
}
```

### Memory Driver

Built-in in-memory cache driver with LRU eviction.

**Features:**
- LRU eviction
- Size limits (items and bytes)
- Metrics collection
- Configurable cleanup

**Example:**
```go
manager.RegisterDriver("memory", memory.NewDriver)
```

### Redis Driver

Redis-based cache driver with serialization support.

**Features:**
- Persistent storage
- Distributed caching
- Serialization (JSON/msgpack)
- Tag support

**Example:**
```go
import "github.com/donnigundala/dg-cache/drivers/redis"

manager.RegisterDriver("redis", redis.NewDriver)
```

## Serialization

### Serializer Interface

```go
type Serializer interface {
    Marshal(v interface{}) ([]byte, error)
    Unmarshal(data []byte, v interface{}) error
    Name() string
}
```

### JSON Serializer

Default serializer using JSON encoding.

**Features:**
- Human-readable
- ~150ns marshal, ~440ns unmarshal
- Good for debugging

**Example:**
```go
Options: map[string]interface{}{
    "serializer": "json",
}
```

### Msgpack Serializer

High-performance binary serializer.

**Features:**
- Compact binary format
- ~210ns marshal, ~172ns unmarshal (2.6x faster!)
- Better for production

**Example:**
```go
Options: map[string]interface{}{
    "serializer": "msgpack",
}
```

## Error Types

### `ErrKeyNotFound`

Returned when a cache key doesn't exist.

**Example:**
```go
val, err := manager.Get(ctx, "key")
if err == cache.ErrKeyNotFound {
    // Key not found
}
```

### `ErrDriverNotFound`

Returned when a driver is not registered.

**Example:**
```go
driver, err := manager.Store("unknown")
if err == cache.ErrDriverNotFound {
    // Driver not found
}
```

## Constants

### Default Values

```go
const (
    DefaultDriver = "memory"
    DefaultTTL    = 0 // No expiration
)
```

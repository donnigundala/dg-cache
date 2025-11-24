# dg-cache

Abstract caching layer for the dg-framework. Provides a unified API for various cache drivers with support for serialization, tagging, atomic operations, and the "Remember" pattern.

## Installation

```bash
go get github.com/donnigundala/dg-cache@v1.1.0
```

## Features

- **Unified API**: Consistent interface across all drivers
- **Serialization**: Automatic marshaling/unmarshaling of complex types (structs, slices, maps)
- **Multiple Serializers**: JSON (default) and Msgpack (2.6x faster unmarshal)
- **Memory Driver**: Production-ready with LRU eviction, size limits, and metrics
- **Multiple Stores**: Support for multiple cache stores in a single application
- **Typed Helpers**: Type-safe retrieval methods (`GetAs`, `GetString`, `GetInt`, etc.)
- **Fluent Interface**: Easy-to-use API inspired by modern frameworks
- **Remember Pattern**: Built-in cache-aside implementation
- **Tagging**: Group related cache items (driver dependent)
- **Atomic Operations**: Increment/Decrement support

## Quick Start

### Basic Usage

```go
import (
    "context"
    "time"
    "github.com/donnigundala/dg-cache"
    "github.com/donnigundala/dg-cache/drivers/memory"
)

func main() {
    // Create manager
    manager, _ := cache.NewManager(cache.DefaultConfig())
    
    // Register memory driver
    manager.RegisterDriver("memory", memory.NewDriver)
    
    ctx := context.Background()
    
    // Store and retrieve strings
    manager.Put(ctx, "name", "John", 10*time.Minute)
    val, _ := manager.Get(ctx, "name")
    name := val.(string) // "John"
}
```

### Caching Complex Types

```go
type User struct {
    ID    int
    Name  string
    Email string
}

// Store any Go type - automatic serialization!
user := User{ID: 1, Name: "John", Email: "john@example.com"}
manager.Put(ctx, "user:1", user, 1*time.Hour)

// Retrieve with type assertion
val, _ := manager.Get(ctx, "user:1")
user = val.(User)

// Or use type-safe helper
var user User
manager.GetAs(ctx, "user:1", &user)
```

### Typed Helpers

```go
// Type-safe retrieval methods
name, err := manager.GetString(ctx, "name")
age, err := manager.GetInt(ctx, "age")
score, err := manager.GetFloat64(ctx, "score")
active, err := manager.GetBool(ctx, "active")

// Generic type-safe method
var config map[string]interface{}
err := manager.GetAs(ctx, "config", &config)
```

### Memory Driver with Limits

```go
manager, _ := cache.NewManager(cache.Config{
    DefaultStore: "memory",
    Stores: map[string]cache.StoreConfig{
        "memory": {
            Driver: "memory",
            Options: map[string]interface{}{
                "max_items":        1000,              // Max 1000 items
                "max_bytes":        10 * 1024 * 1024,  // Max 10MB
                "eviction_policy":  "lru",             // LRU eviction
                "cleanup_interval": 1 * time.Minute,   // Cleanup every minute
                "enable_metrics":   true,              // Enable statistics
            },
        },
    },
})
```

### Redis with Msgpack Serialization

```go
import "github.com/donnigundala/dg-redis"

manager, _ := cache.NewManager(cache.Config{
    DefaultStore: "redis",
    Stores: map[string]cache.StoreConfig{
        "redis": {
            Driver: "redis",
            Options: map[string]interface{}{
                "host":       "localhost",
                "port":       6379,
                "serializer": "msgpack",  // 2.6x faster than JSON!
            },
        },
    },
})

manager.RegisterDriver("redis", redis.NewDriver)
```

### Remember Pattern

Retrieve an item from the cache, or execute the callback and store the result if it doesn't exist.

```go
user, err := manager.Remember(ctx, "user:1", 1*time.Hour, func() (interface{}, error) {
    return db.FindUser(1)
})
```

### Atomic Operations

```go
// Increment
newVal, err := manager.Increment(ctx, "hits", 1)

// Decrement
newVal, err := manager.Decrement(ctx, "hits", 1)
```

### Multiple Stores

```go
// Access specific store
redisStore, err := manager.Store("redis")
redisStore.Put(ctx, "key", "value", 0)

// Access default store
manager.Put(ctx, "key", "value", 0)
```

## Serialization

### Supported Types

- Primitives: `string`, `int`, `float64`, `bool`
- Complex types: `struct`, `slice`, `map`
- Nested structures
- Custom types

### Choosing a Serializer

**JSON** (default):
- Human-readable
- ~150ns marshal, ~440ns unmarshal
- Good for debugging

**Msgpack**:
- Binary format (30-50% smaller)
- ~210ns marshal, ~172ns unmarshal (2.6x faster!)
- Better for production

```go
Options: map[string]interface{}{
    "serializer": "msgpack",  // or "json"
}
```

## Memory Driver Features

### Size Limits

```go
Options: map[string]interface{}{
    "max_items": 1000,              // Maximum number of items
    "max_bytes": 10 * 1024 * 1024,  // Maximum total size (10MB)
}
```

### LRU Eviction

Automatically evicts least recently used items when limits are reached:

```go
Options: map[string]interface{}{
    "eviction_policy": "lru",  // Least Recently Used
}
```

### Metrics

```go
Options: map[string]interface{}{
    "enable_metrics": true,
}

// Get statistics
driver := manager.Store("memory").(*memory.Driver)
stats := driver.Stats()
fmt.Printf("Hit rate: %.2f%%\n", stats.HitRate*100)
fmt.Printf("Items: %d, Bytes: %d\n", stats.ItemCount, stats.BytesUsed)
```

## Creating Custom Drivers

Implement the `cache.Driver` interface:

```go
type Driver interface {
    Get(ctx context.Context, key string) (interface{}, error)
    Put(ctx context.Context, key string, value interface{}, ttl time.Duration) error
    Forget(ctx context.Context, key string) error
    Flush(ctx context.Context) error
    // ... other methods
}
```

## Performance

### Benchmarks

```
BenchmarkJSON_Marshal        7,542,747    152.6 ns/op    128 B/op    2 allocs/op
BenchmarkMsgpack_Marshal     5,384,852    210.9 ns/op    272 B/op    4 allocs/op
BenchmarkJSON_Unmarshal      2,329,303    443.5 ns/op    216 B/op    4 allocs/op
BenchmarkMsgpack_Unmarshal   6,601,837    172.1 ns/op     96 B/op    2 allocs/op
```

**Msgpack is 2.6x faster for unmarshal operations!**

## License

MIT License

## Related Packages

- [dg-redis](https://github.com/donnigundala/dg-redis) - Redis driver for dg-cache
- [dg-core](https://github.com/donnigundala/dg-core) - Core framework


```go
type MyDriver struct {
    // ...
}

func (d *MyDriver) Get(ctx context.Context, key string) (interface{}, error) {
    // ...
}
// ... implement other methods
```

Register it with the manager:

```go
manager.RegisterDriver("my-driver", NewMyDriver)
```

## License

MIT

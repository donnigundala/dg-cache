# dg-cache

Abstract caching layer for the dg-framework. Provides a unified API for various cache drivers with support for serialization, tagging, atomic operations, and the "Remember" pattern.

## Installation

```bash
go get github.com/donnigundala/dg-cache@v1.3.0
```

## Features

- 🚀 **Unified API** - Simple, consistent interface across all drivers
- 🔄 **Multiple Stores** - Use different drivers for different purposes
- 💾 **Built-in Drivers** - Memory (testing) and Redis (production) included
- 🔧 **Extensible** - Easy to add custom drivers
- 📦 **Serialization** - Automatic marshaling/unmarshaling with JSON or Msgpack
- 🏷️ **Tagged Cache** - Group related items with tags (Redis driver)
- ⚡ **Performance** - LRU eviction, metrics, and optimized serialization

## Package Structure

```
dg-cache/
├── config.go              # Configuration structures and validation
├── manager.go             # Cache manager (multi-store orchestration)
├── store.go               # Store, TaggedStore, and Driver interfaces
├── helpers.go             # Typed retrieval helpers (GetString, GetInt, etc.)
├── errors.go              # Custom error types
├── drivers/
│   ├── memory/           # In-memory cache driver
│   │   ├── memory.go     # Core driver implementation
│   │   ├── lru.go        # LRU eviction policy
│   │   ├── metrics.go    # Metrics collection
│   │   └── config.go     # Memory driver configuration
│   └── redis/            # Redis cache driver
│       ├── redis.go      # Core driver implementation
│       ├── tagged.go     # Tagged cache support
│       └── config.go     # Redis driver configuration
├── serializer/
│   ├── serializer.go     # Serializer interface
│   ├── json.go           # JSON serializer
│   └── msgpack.go        # Msgpack serializer
└── docs/
    ├── API.md            # Complete API reference
    ├── SERIALIZATION.md  # Serialization guide
    ├── MEMORY_DRIVER.md  # Memory driver documentation
    └── REDIS_DRIVER.md   # Redis driver documentation
```

## Core Concepts

### Manager
The `Manager` is the central orchestrator that manages multiple cache stores, provides a unified interface across all drivers, handles driver registration, and routes cache operations to the appropriate store.

### Store Interface
Defines the contract that all cache drivers must implement:
- Basic operations: `Get`, `Put`, `Forget`, `Flush`
- Batch operations: `GetMultiple`, `PutMultiple`
- Atomic operations: `Increment`, `Decrement`
- TTL support: `Forever` (no expiration)
- Existence checks: `Has`, `Missing`

### Driver
Extends the Store interface with driver-specific functionality like `Name()` for identification and `Close()` for resource cleanup.

### TaggedStore
Optional interface for drivers that support cache tagging to group related cache items and flush them together. Currently supported by Redis driver only.

## Included Drivers

### Memory Driver (`drivers/memory`)
- In-memory caching for development/testing
- LRU eviction with configurable size limits
- Metrics tracking (hits, misses, evictions)
- Thread-safe operations

### Redis Driver (`drivers/redis`)
- Production-ready Redis caching
- JSON and Msgpack serialization
- Tagged cache support
- Shared client support
- Connection pooling

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
import (
    "github.com/donnigundala/dg-cache"
    "github.com/donnigundala/dg-cache/drivers/redis"
)

// Option 1: Create driver with config
redisDriver, _ := redis.NewDriver(cache.StoreConfig{
    Options: map[string]interface{}{
        "host":       "localhost",
        "port":       6379,
        "serializer": "msgpack", // or "json"
    },
})

// Option 2: Use shared Redis client
client := redis.NewClient(&redis.Options{Addr: "localhost:6379"})
redisDriver := redis.NewDriverWithClient(client, "app")
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

## Documentation

For detailed information, see the comprehensive documentation in the `docs/` directory:

- **[API Reference](docs/API.md)** - Complete API documentation for all packages
- **[Serialization Guide](docs/SERIALIZATION.md)** - Deep dive into JSON and Msgpack serialization
- **[Memory Driver](docs/MEMORY_DRIVER.md)** - In-memory cache with LRU eviction and metrics
- **[Redis Driver](docs/REDIS_DRIVER.md)** - Production-ready Redis caching with tagged cache support

### Version Information

- **Current Version:** v1.3.0
- **Go Version:** 1.21+
- **Test Coverage:** 88%+
- **Status:** Production Ready

## Related Packages

- [dg-core](https://github.com/donnigundala/dg-core) - Core framework for dg-framework
- [dg-database](https://github.com/donnigundala/dg-database) - Database abstraction layer

> **Note:** The `dg-redis` package has been merged into this package as `drivers/redis` in v1.3.0. If you're using the old `dg-redis` package, please migrate to `github.com/donnigundala/dg-cache/drivers/redis`.

## License

MIT License - see [LICENSE](LICENSE) file for details.

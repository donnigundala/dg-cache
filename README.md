# dg-cache

Abstract caching layer for the dg-framework. Provides a unified API for various cache drivers with support for serialization, tagging, atomic operations, and the "Remember" pattern.

## Installation

```bash
go get github.com/donnigundala/dg-cache@v1.0.0
```

## Features

- 🚀 **Unified API** - Simple, consistent interface across all drivers
- 🔄 **Multiple Stores** - Use different drivers for different purposes
- 💾 **Built-in Drivers** - Memory (testing) and Redis (production) included
- 🔧 **Extensible** - Easy to add custom drivers
- 📦 **Serialization** - Automatic marshaling/unmarshaling with JSON or Msgpack
- 🗜️ **Compression** - Transparent Gzip compression for large values
- 📊 **Observability** - Standardized metrics and Prometheus exporter
- 🛡️ **Reliability** - Circuit breaker and enhanced retry logic
- 🏷️ **Tagged Cache** - Group related items with tags (Redis and Memory drivers)
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
Optional interface for drivers that support cache tagging to group related cache items and flush them together. Supported by both Redis and Memory drivers.

## Included Drivers

### Memory Driver (`drivers/memory`)
- In-memory caching for development/testing
- LRU eviction with configurable size limits
- Tagged cache support (v1.6.1)
- Metrics tracking (hits, misses, evictions)
- Thread-safe operations

### Redis Driver (`drivers/redis`)
- Production-ready Redis caching
- JSON and Msgpack serialization
- Tagged cache support
- Shared client support
- Connection pooling

## Quick Start

```go
package main

import (
    "context"
    "log"
    "time"
    "github.com/donnigundala/dg-core/foundation"
    "github.com/donnigundala/dg-cache"
)

func main() {
    app := foundation.New(".")
    
    // Register provider (uses 'cache' key in config)
    app.Register(dgcache.NewCacheServiceProvider(nil))
    
    if err := app.Boot(); err != nil {
        log.Fatal(err)
    }
    
    // Usage
    cacheMgr := dgcache.MustResolve(app)
    ctx := context.Background()
    
    cacheMgr.Put(ctx, "key", "value", 10*time.Minute)
    val, _ := cacheMgr.Get(ctx, "key")
}
```

### Integration via InfrastructureSuite
In your `bootstrap/app.go`, you typically use the declarative suite pattern:

```go
func InfrastructureSuite(workerMode bool) []foundation.ServiceProvider {
	return []foundation.ServiceProvider{
		dgcache.NewCacheServiceProvider(nil),
		// ... other providers
	}
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

## Container Integration (v1.6.0)

As of v1.6.0, dg-cache is fully integrated with the dg-core container system. Named cache stores are automatically registered in the container as `cache.<name>`.

### Access Patterns

#### 1. Direct Resolution
You can resolve specific stores directly from the container:

```go
// Resolve named stores
redisStore, _ := app.Make("cache.redis")
memStore, _ := app.Make("cache.memory")

// Resolve main cache manager
cacheManager, _ := app.Make("cache")
```

#### 2. Helper Functions
Global helper functions provide a more convenient and type-safe way to resolve the cache:

```go
import "github.com/donnigundala/dg-cache"

// Resolve main cache
mgr := cache.MustResolve(app)

// Resolve named store
redis := cache.MustResolveStore(app, "redis")
```

#### 3. Injectable Pattern (Recommended)
The `Injectable` struct simplifies dependency injection in your services:

```go
import (
    "github.com/donnigundala/dg-core/foundation"
    "github.com/donnigundala/dg-cache"
)

type UserService struct {
    inject *cache.Injectable
}

func NewUserService(app foundation.Application) *UserService {
    return &UserService{
        inject: cache.NewInjectable(app),
    }
}

func (s *UserService) CacheUser(ctx context.Context, user *User) {
    // Use default cache store
    s.inject.Cache().Put(ctx, "user:1", user, 0)

    // Use specific store (e.g. redis)
    s.inject.Store("redis").Put(ctx, "user:1", user, 0)
}
```

## Configuration

The plugin uses the `cache` key in your configuration file.

### Configuration Mapping (YAML vs ENV)

| YAML Key | Environment Variable | Default | Description |
| :--- | :--- | :--- | :--- |
| `cache.default_store` | `CACHE_DRIVER` | `memory` | Default store name |
| `cache.prefix` | `CACHE_PREFIX` | `dg_cache` | Global key prefix |
| `cache.stores.<name>.driver` | - | - | `redis`, `memory` |
| `cache.stores.<name>.prefix` | - | - | Store-specific prefix |
| `cache.stores.<name>.connection` | - | `default` | Redis connection name |

### Example YAML

```yaml
cache:
  default_store: redis
  prefix: app_
  stores:
    memory:
      driver: memory
    redis:
      driver: redis
      connection: default
```

### Compression

Enable transparent Gzip compression to save storage space for large values (Redis driver only):

```go
Options: map[string]interface{}{
    "compression": "gzip",
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

## 📊 Observability

`dg-cache` is instrumented with OpenTelemetry metrics. As of v2.0.0, the legacy Prometheus collector has been replaced with native OpenTelemetry instruments.

### Standardized Metrics
The following metrics are automatically collected from all active cache stores via asynchronous observers:

*   `cache_hits_total`: Counter (labels: `cache_store`)
*   `cache_misses_total`: Counter (labels: `cache_store`)
*   `cache_sets_total`: Counter (labels: `cache_store`)
*   `cache_deletes_total`: Counter (labels: `cache_store`)
*   `cache_evictions_total`: Counter (labels: `cache_store`)
*   `cache_items`: Gauge (labels: `cache_store`)
*   `cache_bytes`: Gauge (labels: `cache_store`)

### Configuration
To enable observability, ensure the `dg-observability` plugin is registered and configured:

```yaml
observability:
  enabled: true
  service_name: "my-app"
```

The metrics are automatically registered on application boot. No manual collector registration is required.

## Reliability Features

### Enhanced Retries (Redis)
Configure exponential backoff for Redis connections:

```go
Options: map[string]interface{}{
    "max_retries":       3,
    "min_retry_backoff": 8 * time.Millisecond,
    "max_retry_backoff": 512 * time.Millisecond,
}
```

### Circuit Breaker
Protect your application from cascading cache failures. If the cache becomes unresponsive, the circuit breaker opens and fails fast.

```go
Options: map[string]interface{}{
    "circuit_breaker": map[string]interface{}{
        "enabled":   true,
        "threshold": 5,                 // Fail after 5 errors
        "timeout":   1 * time.Minute,   // Reset after 1 minute
    },
}
```

## Creating Custom Drivers

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

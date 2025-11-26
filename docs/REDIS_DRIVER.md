# Redis Driver Documentation

The Redis driver provides production-ready caching using Redis as the backend storage.

## Import

```go
import "github.com/donnigundala/dg-cache/drivers/redis"
```

## Quick Start

### Option 1: Create Driver with Config

```go
driver, err := redis.NewDriver(cache.StoreConfig{
    Prefix: "myapp",
    Options: map[string]interface{}{
        "host":       "localhost",
        "port":       6379,
        "password":   "", // optional
        "database":   0,
        "pool_size":  10,
        "serializer": "json", // or "msgpack"
    },
})
```

### Option 2: Use Shared Redis Client

```go
import goRedis "github.com/redis/go-redis/v9"

// Create client once
client := goRedis.NewClient(&goRedis.Options{
    Addr: "localhost:6379",
})

// Share with multiple drivers
cacheDriver := redis.NewDriverWithClient(client, "cache")
queueDriver := redis.NewDriverWithClient(client, "queue")
```

## Features

- ✅ Full `cache.Driver` interface implementation
- ✅ JSON and Msgpack serialization
- ✅ Tagged cache support
- ✅ Connection pooling
- ✅ Atomic operations (Increment/Decrement)
- ✅ Shared client support

## Configuration

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| `host` | string | `localhost` | Redis server host |
| `port` | int | `6379` | Redis server port |
| `password` | string | `""` | Redis password |
| `database` | int | `0` | Redis database number |
| `pool_size` | int | `10` | Connection pool size |
| `serializer` | string | `json` | Serializer (`json` or `msgpack`) |

## Tagged Cache

```go
// Create tagged cache
tagged := driver.Tags("users", "active")

// Store with tags
tagged.Put(ctx, "user:1", user, 1*time.Hour)

// Flush all keys with these tags
driver.FlushTags(ctx, "users")
```

## Typed Helpers

The Redis driver supports all typed helper methods for type-safe retrieval:

```go
import (
    "context"
    "github.com/donnigundala/dg-cache"
    "github.com/donnigundala/dg-cache/drivers/redis"
)

// Setup
manager, _ := cache.NewManager(cache.Config{
    DefaultStore: "redis",
    Stores: map[string]cache.StoreConfig{
        "redis": {
            Driver: "redis",
            Options: map[string]interface{}{
                "host": "localhost",
                "port": 6379,
            },
        },
    },
})
manager.RegisterDriver("redis", redis.NewDriver)

ctx := context.Background()

// String retrieval
name, err := manager.GetString(ctx, "user:name")

// Integer retrieval
age, err := manager.GetInt(ctx, "user:age")

// Float retrieval
score, err := manager.GetFloat64(ctx, "user:score")

// Boolean retrieval
active, err := manager.GetBool(ctx, "user:active")

// Generic type-safe unmarshaling
type User struct {
    ID    int
    Name  string
    Email string
}

var user User
err := manager.GetAs(ctx, "user:1", &user)
```

## Manager Integration

### Basic Setup with Manager

```go
import (
    "github.com/donnigundala/dg-cache"
    "github.com/donnigundala/dg-cache/drivers/redis"
)

config := cache.Config{
    DefaultStore: "redis",
    Prefix:       "myapp",
    Stores: map[string]cache.StoreConfig{
        "redis": {
            Driver: "redis",
            Prefix: "cache",
            Options: map[string]interface{}{
                "host":       "localhost",
                "port":       6379,
                "serializer": "msgpack",
            },
        },
    },
}

manager, err := cache.NewManager(config)
if err != nil {
    panic(err)
}

// Register Redis driver
manager.RegisterDriver("redis", redis.NewDriver)

// Use manager methods (delegates to Redis)
manager.Put(ctx, "key", "value", 10*time.Minute)
val, _ := manager.Get(ctx, "key")
```

### Multi-Store Configuration

Use Redis alongside memory cache:

```go
import (
    "github.com/donnigundala/dg-cache"
    "github.com/donnigundala/dg-cache/drivers/memory"
    "github.com/donnigundala/dg-cache/drivers/redis"
)

config := cache.Config{
    DefaultStore: "redis",
    Stores: map[string]cache.StoreConfig{
        "redis": {
            Driver: "redis",
            Options: map[string]interface{}{
                "host": "localhost",
                "port": 6379,
            },
        },
        "memory": {
            Driver: "memory",
            Options: map[string]interface{}{
                "max_items": 1000,
            },
        },
    },
}

manager, _ := cache.NewManager(config)
manager.RegisterDriver("redis", redis.NewDriver)
manager.RegisterDriver("memory", memory.NewDriver)

// Use Redis (default store)
manager.Put(ctx, "session:abc", sessionData, 1*time.Hour)

// Use memory for temporary data
memStore, _ := manager.Store("memory")
memStore.Put(ctx, "temp:xyz", tempData, 1*time.Minute)
```

### Remember Pattern with Redis

```go
// Cache-aside pattern with Redis
user, err := manager.Remember(ctx, "user:1", 1*time.Hour, func() (interface{}, error) {
    // This only executes on cache miss
    return database.FindUser(1)
})
```

## Performance

- **Msgpack**: 2.6x faster unmarshaling than JSON (172ns vs 443ns)
- **Connection Pooling**: Reuses connections efficiently
- **Pipelining**: Batch operations use Redis pipelines

### Serialization Benchmarks

```
JSON Unmarshal:    443.5 ns/op    216 B/op    4 allocs/op
Msgpack Unmarshal: 172.1 ns/op     96 B/op    2 allocs/op
```

**Recommendation:** Use Msgpack serializer in production for better performance.

## Migration from dg-redis

The Redis driver was previously a separate `dg-redis` package. It has been merged into dg-cache for better architecture.

**Before:**
```go
import "github.com/donnigundala/dg-redis"

driver, _ := redis.NewDriver(config)
```

**After:**
```go
import "github.com/donnigundala/dg-cache/drivers/redis"

driver, _ := redis.NewDriver(config)
```

The API is identical, only the import path changed.

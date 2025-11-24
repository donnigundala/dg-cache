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
driver.Tags("users").Flush()
```

## Performance

- **Msgpack**: 6.7x faster unmarshaling than JSON
- **Connection Pooling**: Reuses connections efficiently
- **Pipelining**: Batch operations use Redis pipelines

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

# dgcore-cache

Abstract caching layer for the dgcore framework. Provides a unified API for various cache drivers with support for tagging, atomic operations, and the "Remember" pattern.

## Installation

```bash
go get github.com/donnigundala/dg-cache
```

## Features

- **Unified API**: Consistent interface across all drivers
- **Multiple Stores**: Support for multiple cache stores in a single application
- **Fluent Interface**: Easy-to-use API inspired by modern frameworks
- **Remember Pattern**: Built-in cache-aside implementation
- **Tagging**: Group related cache items (driver dependent)
- **Atomic Operations**: Increment/Decrement support
- **In-Memory Driver**: Built-in driver for testing and local development

## Usage

### Initialization

```go
import (
    "github.com/donnigundala/dgcore-cache"
    "github.com/donnigundala/dgcore-cache/drivers/memory"
)

func main() {
    // Configure
    cfg := cache.DefaultConfig()
    
    // Create manager
    manager, err := cache.NewManager(cfg)
    if err != nil {
        panic(err)
    }

    // Register drivers
    manager.RegisterDriver("memory", memory.NewDriver)
}
```

### Basic Operations

```go
ctx := context.Background()

// Store a value
err := manager.Put(ctx, "user:1", user, 10*time.Minute)

// Retrieve a value
val, err := manager.Get(ctx, "user:1")

// Check existence
exists, err := manager.Has(ctx, "user:1")

// Remove a value
err := manager.Forget(ctx, "user:1")

// Retrieve and delete
val, err := manager.Pull(ctx, "temp_key")
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

## Creating Custom Drivers

Implement the `cache.Driver` interface to create a custom driver.

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

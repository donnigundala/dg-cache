# Memory Driver Guide

This guide covers the in-memory cache driver with LRU eviction, size limits, and metrics.

## Table of Contents
- [Overview](#overview)
- [Features](#features)
- [Configuration](#configuration)
- [Size Limits](#size-limits)
- [LRU Eviction](#lru-eviction)
- [Metrics](#metrics)
- [Best Practices](#best-practices)

## Overview

The memory driver is a built-in cache implementation that stores data in your application's RAM. It's perfect for testing, development, and single-server deployments.

### When to Use

**✅ Good for:**
- Unit testing (no external dependencies)
- Development/local testing
- Session data (single server)
- Temporary computations
- Small datasets that fit in RAM

**❌ Not good for:**
- Production (multi-server deployments)
- Data that must survive restarts
- Sharing cache between app instances
- Large datasets (memory constraints)

## Features

### Production-Ready Features

- **LRU Eviction**: Automatically evicts least recently used items
- **Size Limits**: Configurable max items and max bytes
- **Metrics Collection**: Track hits, misses, evictions, and size
- **Configurable Cleanup**: Customize expired item removal interval
- **Thread-Safe**: Safe for concurrent access
- **Zero Dependencies**: No external services required

## Configuration

### Basic Configuration

```go
import (
    "github.com/donnigundala/dg-cache"
    "github.com/donnigundala/dg-cache/drivers/memory"
)

config := cache.Config{
    DefaultStore: "memory",
    Stores: map[string]cache.StoreConfig{
        "memory": {
            Driver: "memory",
        },
    },
}

manager, _ := cache.NewManager(config)
manager.RegisterDriver("memory", memory.NewDriver)
```

### Full Configuration

```go
config := cache.Config{
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
}
```

## Size Limits

### Item Count Limit

Limit the maximum number of items in the cache:

```go
Options: map[string]interface{}{
    "max_items": 1000, // Maximum 1000 items
}
```

**Behavior:**
- When limit is reached, LRU item is evicted before adding new item
- 0 = unlimited (default)

**Example:**
```go
// Configure with 3 item limit
Options: map[string]interface{}{
    "max_items": 3,
}

// Add 3 items
cache.Put(ctx, "key1", "value1", 0)
cache.Put(ctx, "key2", "value2", 0)
cache.Put(ctx, "key3", "value3", 0)

// Add 4th item - evicts least recently used (key1)
cache.Put(ctx, "key4", "value4", 0)

// key1 is now evicted
_, err := cache.Get(ctx, "key1") // Returns ErrKeyNotFound
```

### Byte Size Limit

Limit the total size of cached data:

```go
Options: map[string]interface{}{
    "max_bytes": 10 * 1024 * 1024, // Maximum 10MB
}
```

**Behavior:**
- Tracks estimated size of all cached values
- Evicts LRU items until new item fits
- 0 = unlimited (default)

**Size Estimation:**
- `string` / `[]byte`: actual length
- Numeric types: 8 bytes
- `bool`: 1 byte
- Complex types: 64 bytes (default estimate)

**Example:**
```go
// Configure with 100 byte limit
Options: map[string]interface{}{
    "max_bytes": 100,
}

// Add items totaling 90 bytes
cache.Put(ctx, "key1", "12345678901234567890", 0) // 20 bytes
cache.Put(ctx, "key2", "12345678901234567890", 0) // 20 bytes
cache.Put(ctx, "key3", "12345678901234567890123456789012345678901234567890", 0) // 50 bytes

// Add large item - triggers eviction
cache.Put(ctx, "key4", "123456789012345678901234567890", 0) // 30 bytes
// Evicts key1 and key2 to make room
```

## LRU Eviction

### How It Works

The Least Recently Used (LRU) eviction policy tracks access order using a doubly-linked list:

1. **Access**: When an item is accessed (Get/Put), it moves to the front
2. **Eviction**: When limits are reached, items at the back are evicted first
3. **O(1) Operations**: All operations are constant time

**Example:**
```go
// Configure LRU eviction
Options: map[string]interface{}{
    "max_items":       3,
    "eviction_policy": "lru",
}

// Add 3 items
cache.Put(ctx, "key1", "value1", 0) // Order: key1
cache.Put(ctx, "key2", "value2", 0) // Order: key2, key1
cache.Put(ctx, "key3", "value3", 0) // Order: key3, key2, key1

// Access key1 (moves to front)
cache.Get(ctx, "key1") // Order: key1, key3, key2

// Add key4 - evicts key2 (least recently used)
cache.Put(ctx, "key4", "value4", 0) // Order: key4, key1, key3
```

### Eviction Scenarios

**Scenario 1: Max Items Reached**
```go
Options: map[string]interface{}{
    "max_items": 100,
}

// When 101st item is added, LRU item is evicted
```

**Scenario 2: Max Bytes Exceeded**
```go
Options: map[string]interface{}{
    "max_bytes": 1024 * 1024, // 1MB
}

// When adding item would exceed 1MB, evict LRU items until it fits
```

**Scenario 3: Both Limits**
```go
Options: map[string]interface{}{
    "max_items": 1000,
    "max_bytes": 10 * 1024 * 1024,
}

// Whichever limit is hit first triggers eviction
```

## Metrics

### Enabling Metrics

```go
Options: map[string]interface{}{
    "enable_metrics": true,
}
```

### Accessing Metrics

```go
// Get driver instance
driver := manager.Store("memory").(*memory.Driver)

// Get statistics
stats := driver.Stats()

fmt.Printf("Hit rate: %.2f%%\n", stats.HitRate*100)
fmt.Printf("Items: %d\n", stats.ItemCount)
fmt.Printf("Bytes: %d\n", stats.BytesUsed)
fmt.Printf("Hits: %d\n", stats.Hits)
fmt.Printf("Misses: %d\n", stats.Misses)
fmt.Printf("Evictions: %d\n", stats.Evictions)
```

### Available Metrics

```go
type Stats struct {
    Hits       int64   // Number of cache hits
    Misses     int64   // Number of cache misses
    HitRate    float64 // Hit rate (hits / (hits + misses))
    Sets       int64   // Number of Set operations
    Deletes    int64   // Number of Delete operations
    Evictions  int64   // Number of evictions
    ItemCount  int     // Current number of items
    BytesUsed  int64   // Current bytes used
}
```

### Monitoring Example

```go
// Periodic monitoring
ticker := time.NewTicker(1 * time.Minute)
go func() {
    for range ticker.C {
        stats := driver.Stats()
        log.Printf("Cache stats: hit_rate=%.2f%% items=%d bytes=%d evictions=%d",
            stats.HitRate*100, stats.ItemCount, stats.BytesUsed, stats.Evictions)
    }
}()
```

## Best Practices

### 1. Set Appropriate Limits

```go
// Development: Generous limits
Options: map[string]interface{}{
    "max_items": 10000,
    "max_bytes": 100 * 1024 * 1024, // 100MB
}

// Production: Conservative limits
Options: map[string]interface{}{
    "max_items": 1000,
    "max_bytes": 10 * 1024 * 1024, // 10MB
}
```

### 2. Monitor Metrics

```go
// Enable metrics in production
Options: map[string]interface{}{
    "enable_metrics": true,
}

// Alert on low hit rate
stats := driver.Stats()
if stats.HitRate < 0.7 { // Less than 70%
    log.Warn("Low cache hit rate")
}
```

### 3. Use TTL for Expiration

```go
// Short-lived data
cache.Put(ctx, "session:"+sessionID, session, 15*time.Minute)

// Long-lived data
cache.Put(ctx, "config", config, 24*time.Hour)

// Permanent data (until evicted)
cache.Put(ctx, "static_data", data, 0)
```

### 4. Tune Cleanup Interval

```go
// Frequent cleanup (more CPU, less memory)
Options: map[string]interface{}{
    "cleanup_interval": 30 * time.Second,
}

// Infrequent cleanup (less CPU, more memory)
Options: map[string]interface{}{
    "cleanup_interval": 5 * time.Minute,
}
```

### 5. Test Eviction Behavior

```go
func TestCacheEviction(t *testing.T) {
    config := cache.StoreConfig{
        Driver: "memory",
        Options: map[string]interface{}{
            "max_items":      3,
            "enable_metrics": true,
        },
    }
    
    driver, _ := memory.NewDriver(config)
    ctx := context.Background()
    
    // Fill cache
    driver.Put(ctx, "key1", "value1", 0)
    driver.Put(ctx, "key2", "value2", 0)
    driver.Put(ctx, "key3", "value3", 0)
    
    // Trigger eviction
    driver.Put(ctx, "key4", "value4", 0)
    
    // Verify eviction
    _, err := driver.Get(ctx, "key1")
    if err != cache.ErrKeyNotFound {
        t.Error("key1 should have been evicted")
    }
}
```

## Common Patterns

### Session Cache

```go
// Configure for sessions
config := cache.Config{
    DefaultStore: "memory",
    Stores: map[string]cache.StoreConfig{
        "memory": {
            Driver: "memory",
            Options: map[string]interface{}{
                "max_items":        10000,
                "cleanup_interval": 1 * time.Minute,
            },
        },
    },
}

// Store session
cache.Put(ctx, "session:"+sessionID, session, 30*time.Minute)
```

### Computation Cache

```go
// Cache expensive computations
result, err := cache.Remember(ctx, "computation:"+key, 1*time.Hour, func() (interface{}, error) {
    return expensiveComputation(key)
})
```

### API Response Cache

```go
// Cache API responses
response, err := cache.Remember(ctx, "api:users:list", 5*time.Minute, func() (interface{}, error) {
    return api.FetchUsers()
})
```

## Performance

### Characteristics

- **Get**: O(1) hash lookup + O(1) LRU update = **O(1)**
- **Put**: O(1) eviction check + O(1) insert + O(1) LRU update = **O(1)**
- **Eviction**: O(1) LRU removal + O(1) hash deletion = **O(1)**
- **Memory**: O(n) for items + O(n) for LRU nodes = **O(n)**

### Benchmarks

```
BenchmarkMemory_Get     50,000,000    25 ns/op
BenchmarkMemory_Put     30,000,000    40 ns/op
BenchmarkMemory_Evict   20,000,000    60 ns/op
```

## Troubleshooting

### High Eviction Rate

**Problem**: Too many evictions

**Solutions:**
1. Increase `max_items` or `max_bytes`
2. Reduce TTL for less important data
3. Use Redis for larger datasets

### Low Hit Rate

**Problem**: Cache not effective

**Solutions:**
1. Increase cache size
2. Review access patterns
3. Adjust TTL values

### Memory Growth

**Problem**: Memory usage growing

**Solutions:**
1. Set `max_bytes` limit
2. Reduce `cleanup_interval`
3. Lower TTL values

## Migration from Redis

```go
// Before: Redis
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
    },
}

// After: Memory (for testing)
config := cache.Config{
    DefaultStore: "memory",
    Stores: map[string]cache.StoreConfig{
        "memory": {
            Driver: "memory",
            Options: map[string]interface{}{
                "max_items": 1000,
            },
        },
    },
}

// Code remains the same!
cache.Put(ctx, "key", "value", 0)
```

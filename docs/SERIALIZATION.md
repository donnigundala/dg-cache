# Serialization Guide

This guide covers serialization in dg-cache, enabling you to cache complex Go types.

## Table of Contents
- [Overview](#overview)
- [Supported Types](#supported-types)
- [Serializers](#serializers)
- [Configuration](#configuration)
- [Best Practices](#best-practices)
- [Performance](#performance)

## Overview

dg-cache supports automatic serialization of complex Go types, allowing you to cache structs, slices, maps, and more without manual marshaling.

### Before Serialization

```go
// Manual JSON marshaling
user := User{ID: 1, Name: "John"}
jsonBytes, _ := json.Marshal(user)
cache.Put(ctx, "user:1", string(jsonBytes), 0)

// Manual unmarshaling
val, _ := cache.Get(ctx, "user:1")
json.Unmarshal([]byte(val.(string)), &user)
```

### With Serialization

```go
// Automatic serialization!
user := User{ID: 1, Name: "John"}
cache.Put(ctx, "user:1", user, 0)

// Automatic deserialization!
val, _ := cache.Get(ctx, "user:1")
user = val.(User)
```

## Supported Types

### Primitive Types

```go
// Strings
cache.Put(ctx, "name", "John", 0)
name, _ := cache.GetString(ctx, "name")

// Numbers
cache.Put(ctx, "age", 30, 0)
age, _ := cache.GetInt(ctx, "age")

cache.Put(ctx, "price", 19.99, 0)
price, _ := cache.GetFloat64(ctx, "price")

// Booleans
cache.Put(ctx, "active", true, 0)
active, _ := cache.GetBool(ctx, "active")
```

### Structs

```go
type User struct {
    ID    int
    Name  string
    Email string
}

user := User{ID: 1, Name: "John", Email: "john@example.com"}
cache.Put(ctx, "user:1", user, 1*time.Hour)

var cached User
cache.GetAs(ctx, "user:1", &cached)
```

### Nested Structs

```go
type Address struct {
    Street string
    City   string
}

type Person struct {
    Name    string
    Age     int
    Address Address
}

person := Person{
    Name: "Alice",
    Age:  25,
    Address: Address{
        Street: "123 Main St",
        City:   "Springfield",
    },
}

cache.Put(ctx, "person:1", person, 0)
```

### Slices

```go
// Slice of primitives
items := []int{1, 2, 3, 4, 5}
cache.Put(ctx, "items", items, 0)

// Slice of structs
users := []User{
    {ID: 1, Name: "John"},
    {ID: 2, Name: "Jane"},
}
cache.Put(ctx, "users", users, 0)
```

### Maps

```go
// Map of primitives
config := map[string]string{
    "host": "localhost",
    "port": "8080",
}
cache.Put(ctx, "config", config, 0)

// Map of interfaces
data := map[string]interface{}{
    "name":  "John",
    "age":   30,
    "admin": true,
}
cache.Put(ctx, "data", data, 0)
```

### Custom Types

```go
type UserID int
type Email string

type CustomUser struct {
    ID    UserID
    Email Email
}

user := CustomUser{
    ID:    UserID(1),
    Email: Email("john@example.com"),
}
cache.Put(ctx, "custom_user", user, 0)
```

## Serializers

### JSON Serializer (Default)

Human-readable serialization using Go's `encoding/json`.

**Pros:**
- Human-readable
- Easy to debug
- Wide compatibility
- ~150ns marshal, ~440ns unmarshal

**Cons:**
- Larger payload size
- Slower than binary formats

**Configuration:**
```go
config := cache.Config{
    Stores: map[string]cache.StoreConfig{
        "redis": {
            Driver: "redis",
            Options: map[string]interface{}{
                "serializer": "json", // or omit for default
            },
        },
    },
}
```

**Example Output:**
```json
{
  "type": "main.User",
  "value": {
    "ID": 1,
    "Name": "John",
    "Email": "john@example.com"
  }
}
```

### Msgpack Serializer

High-performance binary serialization using MessagePack.

**Pros:**
- Compact binary format (30-50% smaller)
- **2.6x faster unmarshal** than JSON
- Better for production
- ~210ns marshal, ~172ns unmarshal

**Cons:**
- Not human-readable
- Requires msgpack dependency

**Configuration:**
```go
config := cache.Config{
    Stores: map[string]cache.StoreConfig{
        "redis": {
            Driver: "redis",
            Options: map[string]interface{}{
                "serializer": "msgpack",
            },
        },
    },
}
```

## Configuration

### Redis Driver

```go
import (
    "github.com/donnigundala/dg-cache"
    "github.com/donnigundala/dg-redis"
)

// JSON serializer (default)
config := cache.Config{
    DefaultStore: "redis",
    Stores: map[string]cache.StoreConfig{
        "redis": {
            Driver: "redis",
            Options: map[string]interface{}{
                "host":       "localhost",
                "port":       6379,
                "serializer": "json",
            },
        },
    },
}

// Msgpack serializer (faster)
config := cache.Config{
    DefaultStore: "redis",
    Stores: map[string]cache.StoreConfig{
        "redis": {
            Driver: "redis",
            Options: map[string]interface{}{
                "host":       "localhost",
                "port":       6379,
                "serializer": "msgpack",
            },
        },
    },
}
```

### Memory Driver

The memory driver doesn't use serialization (stores values directly in memory).

## Best Practices

### 1. Use Type-Safe Helpers

```go
// Good - Type-safe
var user User
if err := cache.GetAs(ctx, "user:1", &user); err != nil {
    // Handle error gracefully
}

// Avoid - Type assertion can panic
val, _ := cache.Get(ctx, "user:1")
user := val.(User) // Panic if wrong type!
```

### 2. Handle Serialization Errors

```go
err := cache.Put(ctx, "key", value, 0)
if err != nil {
    log.Printf("Failed to cache value: %v", err)
    // Continue without caching
}
```

### 3. Use Appropriate Serializer

```go
// Development/Debugging: Use JSON
Options: map[string]interface{}{
    "serializer": "json",
}

// Production: Use Msgpack
Options: map[string]interface{}{
    "serializer": "msgpack",
}
```

### 4. Cache Computed Values

```go
type UserStats struct {
    TotalPosts    int
    TotalComments int
    LastActive    time.Time
}

stats, err := cache.Remember(ctx, "user:1:stats", 5*time.Minute, func() (interface{}, error) {
    // Expensive computation
    return computeUserStats(userID)
})
```

### 5. Version Your Cached Structures

```go
type User struct {
    Version int    `json:"version"` // Add version field
    ID      int    `json:"id"`
    Name    string `json:"name"`
}

// When retrieving
var user User
if err := cache.GetAs(ctx, "user:1", &user); err == nil {
    if user.Version != CurrentUserVersion {
        // Invalidate and reload
        cache.Forget(ctx, "user:1")
    }
}
```

### 6. Use Prefixes for Namespacing

```go
// Group related cache keys
cache.Put(ctx, "user:1:profile", profile, 0)
cache.Put(ctx, "user:1:settings", settings, 0)
cache.Put(ctx, "user:1:stats", stats, 0)
```

## Performance

### Benchmarks

```
BenchmarkJSON_Marshal        7,542,747    152.6 ns/op    128 B/op    2 allocs/op
BenchmarkMsgpack_Marshal     5,384,852    210.9 ns/op    272 B/op    4 allocs/op
BenchmarkJSON_Unmarshal      2,329,303    443.5 ns/op    216 B/op    4 allocs/op
BenchmarkMsgpack_Unmarshal   6,601,837    172.1 ns/op     96 B/op    2 allocs/op
```

### Key Findings

1. **Msgpack unmarshal is 2.6x faster** (172ns vs 443ns)
2. **Msgpack uses less memory** for unmarshal (96B vs 216B)
3. **JSON marshal is slightly faster** (152ns vs 210ns)
4. **Both are extremely fast** for typical cache operations

### Size Comparison

```go
user := User{ID: 1, Name: "John Doe", Email: "john@example.com"}

// JSON: ~120 bytes
{
  "type": "main.User",
  "value": {"ID":1,"Name":"John Doe","Email":"john@example.com"}
}

// Msgpack: ~80 bytes (33% smaller)
// Binary format, not human-readable
```

### When to Use Each

**Use JSON when:**
- Debugging cache issues
- Need human-readable cache
- Compatibility is important
- Cache size is not a concern

**Use Msgpack when:**
- Production environment
- Performance is critical
- Cache size matters
- High throughput required

## Common Patterns

### Caching API Responses

```go
type APIResponse struct {
    Data      interface{}
    Timestamp time.Time
    CacheKey  string
}

response, err := cache.Remember(ctx, "api:users:list", 5*time.Minute, func() (interface{}, error) {
    data, err := api.FetchUsers()
    if err != nil {
        return nil, err
    }
    return APIResponse{
        Data:      data,
        Timestamp: time.Now(),
        CacheKey:  "api:users:list",
    }, nil
})
```

### Caching Database Queries

```go
type QueryResult struct {
    Users []User
    Total int
}

result, err := cache.Remember(ctx, "query:active_users", 10*time.Minute, func() (interface{}, error) {
    var users []User
    db.Where("active = ?", true).Find(&users)
    return QueryResult{
        Users: users,
        Total: len(users),
    }, nil
})
```

### Caching with Fallback

```go
var user User
err := cache.GetAs(ctx, "user:1", &user)
if err == cache.ErrKeyNotFound {
    // Load from database
    user, err = db.FindUser(1)
    if err != nil {
        return err
    }
    // Cache for next time
    cache.Put(ctx, "user:1", user, 1*time.Hour)
}
```

## Troubleshooting

### Type Mismatch Errors

```go
// Problem: Cached as User, trying to get as Admin
cache.Put(ctx, "user:1", User{}, 0)
var admin Admin
err := cache.GetAs(ctx, "user:1", &admin) // Error!

// Solution: Use correct type or clear cache
cache.Forget(ctx, "user:1")
```

### Serialization Failures

```go
// Problem: Type contains unexported fields
type User struct {
    id   int    // unexported!
    Name string
}

// Solution: Export fields or use tags
type User struct {
    ID   int    `json:"id"`
    Name string `json:"name"`
}
```

### Large Payloads

```go
// Problem: Caching very large objects
type HugeData struct {
    Data [1000000]int
}

// Solution: Cache only what you need
type CompactData struct {
    Summary Stats
    TopN    []int
}
```

## Migration Guide

### From Manual Serialization

```go
// Before
userJSON, _ := json.Marshal(user)
cache.Put(ctx, "user:1", string(userJSON), 0)

val, _ := cache.Get(ctx, "user:1")
json.Unmarshal([]byte(val.(string)), &user)

// After
cache.Put(ctx, "user:1", user, 0)
cache.GetAs(ctx, "user:1", &user)
```

### From String-Only Caching

```go
// Before
cache.Put(ctx, "user_id", "123", 0)
idStr, _ := cache.GetString(ctx, "user_id")
id, _ := strconv.Atoi(idStr)

// After
cache.Put(ctx, "user_id", 123, 0)
id, _ := cache.GetInt(ctx, "user_id")
```

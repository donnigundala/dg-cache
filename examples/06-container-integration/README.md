# Container Integration Example

This example demonstrates the new container integration features introduced in **dg-cache v1.6.0**.

## Features Demonstrated

1. **Auto-Registration**: Named cache stores are automatically registered in the container (`cache.sessions`, `cache.api_cache`).
2. **Direct Resolution**: Resolving stores via `app.Make("cache.sessions")`.
3. **Helper Functions**: Using `cache.Resolve()` and `cache.ResolveStore()`.
4. **Injectable Pattern**: Clean dependency injection in services using `cache.Injectable`.

## Running the Example

```bash
go run main.go
```

## Expected Output

```
🚀 Cache Container Integration Example
----------------------------------------

1. Direct Resolution:
   Main cache resolved: *cache.Manager
   Sessions store resolved: *memory.Driver
   API store resolved: *memory.Driver

2. Helper Functions:
   MustResolve success: *cache.Manager
   MustResolveStore('sessions') success: *memory.Driver

3. Injectable Pattern (Service):
   - Data cached in default store
   - Data cached in 'sessions' store
   Main Store (memory) has 'user:1': true
   Sessions Store has 'user:1': true
```

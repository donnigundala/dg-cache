# Container Integration Example

This example demonstrates the core container integration patterns for accessing cache stores in the DG Framework.

## Features Demonstrated

1. **Auto-Registration**: Named cache stores are automatically registered in the container (`cache.sessions`, `cache.api_cache`).
2. **Direct Resolution**: Resolving stores via `app.Make("cache.sessions")`.
3. **Helper Functions**: Using `cache.Resolve()` and `cache.ResolveStore()`.
4. **Injectable Pattern**: Clean dependency injection in services using `cache.Injectable`.
5. **Phase 6 Service Provider**: Standardized registration pattern.

## Running the Example

```bash
go run main.go
```

## Key Integration Patterns

### Registration

The provider handles booting and registration of all stores. In this example, we set the config manually:

```go
config := cache.Config{...}
provider := cache.NewCacheServiceProvider(nil)
provider.Config = config
app.Register(provider)
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

### Pattern 1: Direct Resolution
```go
sessionsStore, _ := app.Make("cache.sessions")
```

### Pattern 2: Helper Functions
```go
mgr := cache.MustResolve(app)
sessions := cache.MustResolveStore(app, "sessions")
```

### Pattern 3: Injectable (Recommended)
```go
type UserService struct {
    inject *cache.Injectable
}

func NewUserService(app *foundation.Application) *UserService {
    return &UserService{
        inject: cache.NewInjectable(app),
    }
}
```

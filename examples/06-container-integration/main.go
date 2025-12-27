package main

import (
	"context"
	"fmt"
	"log"
	"time"

	cache "github.com/donnigundala/dg-cache"
	"github.com/donnigundala/dg-core/foundation"
)

func main() {
	// 1. Create a new application
	app := foundation.New(".")

	// 2. Configure Cache Service
	config := cache.Config{
		DefaultStore: "memory",
		Stores: map[string]cache.StoreConfig{
			"memory": {
				Driver: "memory",
				Options: map[string]interface{}{
					"max_items": 100,
				},
			},
			"sessions": {
				Driver: "memory",
				Options: map[string]interface{}{
					"max_items": 1000,
				},
			},
			"api_cache": {
				Driver: "memory",
				Options: map[string]interface{}{
					"max_items": 500,
				},
			},
		},
	}

	// 3. Register Provider
	provider := cache.NewCacheServiceProvider(nil)
	provider.Config = config // Manually set config for this example

	app.Register(provider)

	// 4. Boot Application (resolves dependencies)
	if err := app.Boot(); err != nil {
		log.Fatalf("Failed to boot application: %v", err)
	}

	fmt.Println("🚀 Cache Container Integration Example")
	fmt.Println("----------------------------------------")

	ctx := context.Background()

	// Pattern 1: Direct Resolution
	fmt.Println("\n1. Direct Resolution:")
	mainCache, _ := app.Make("cache")
	fmt.Printf("   Main cache resolved: %T\n", mainCache)

	sessionsStore, _ := app.Make("cache.sessions")
	fmt.Printf("   Sessions store resolved: %T\n", sessionsStore)

	apiStore, _ := app.Make("cache.api_cache")
	fmt.Printf("   API store resolved: %T\n", apiStore)

	// Pattern 2: Helper Functions
	fmt.Println("\n2. Helper Functions:")
	mgr := cache.MustResolve(app)
	fmt.Printf("   MustResolve success: %T\n", mgr)

	store := cache.MustResolveStore(app, "sessions")
	fmt.Printf("   MustResolveStore('sessions') success: %T\n", store)

	// Pattern 3: Injectable Pattern (Service)
	fmt.Println("\n3. Injectable Pattern (Service):")

	userService := NewUserService(app)
	userService.CacheUserData(ctx, 1, "John Doe")

	// Verify data in different stores
	hasMain, _ := mgr.Has(ctx, "user:1")
	hasSessions, _ := cache.MustResolveStore(app, "sessions").Has(ctx, "user:1")

	fmt.Printf("   Main Store (memory) has 'user:1': %v\n", hasMain)
	fmt.Printf("   Sessions Store has 'user:1': %v\n", hasSessions)
}

// UserService demonstrates dependency injection
type UserService struct {
	inject *cache.Injectable
}

func NewUserService(app *foundation.Application) *UserService {
	return &UserService{
		inject: cache.NewInjectable(app),
	}
}

func (s *UserService) CacheUserData(ctx context.Context, id int, name string) {
	key := fmt.Sprintf("user:%d", id)

	// Use default store (main memory)
	err := s.inject.Cache().Put(ctx, key, name, 1*time.Hour)
	if err != nil {
		fmt.Printf("   Error putting to default store: %v\n", err)
	} else {
		fmt.Println("   - Data cached in default store")
	}

	// Use specific store (sessions)
	err = s.inject.Store("sessions").Put(ctx, key, name, 30*time.Minute)
	if err != nil {
		fmt.Printf("   Error putting to sessions store: %v\n", err)
	} else {
		fmt.Println("   - Data cached in 'sessions' store")
	}
}

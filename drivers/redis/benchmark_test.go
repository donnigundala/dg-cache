package redis

import (
	"context"
	"testing"
	"time"

	cache "github.com/donnigundala/dg-cache"
)

// BenchmarkJSON_Put benchmarks Put operation with JSON serializer
func BenchmarkJSON_Put(b *testing.B) {
	config := cache.StoreConfig{
		Driver: "redis",
		Options: map[string]interface{}{
			"host":       "localhost",
			"port":       6379,
			"database":   15,
			"serializer": "json",
		},
	}

	driver, err := NewDriver(config)
	if err != nil {
		b.Skipf("Skipping benchmark: Redis not available: %v", err)
	}
	defer driver.Close()

	redisDriver := driver.(*Driver)
	ctx := context.Background()

	type User struct {
		ID    int
		Name  string
		Email string
	}

	user := User{ID: 1, Name: "John Doe", Email: "john@example.com"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = redisDriver.Put(ctx, "bench:user", user, 1*time.Minute)
	}
}

// BenchmarkMsgpack_Put benchmarks Put operation with msgpack serializer
func BenchmarkMsgpack_Put(b *testing.B) {
	config := cache.StoreConfig{
		Driver: "redis",
		Options: map[string]interface{}{
			"host":       "localhost",
			"port":       6379,
			"database":   15,
			"serializer": "msgpack",
		},
	}

	driver, err := NewDriver(config)
	if err != nil {
		b.Skipf("Skipping benchmark: Redis not available: %v", err)
	}
	defer driver.Close()

	redisDriver := driver.(*Driver)
	ctx := context.Background()

	type User struct {
		ID    int
		Name  string
		Email string
	}

	user := User{ID: 1, Name: "John Doe", Email: "john@example.com"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = redisDriver.Put(ctx, "bench:user", user, 1*time.Minute)
	}
}

// BenchmarkJSON_Get benchmarks Get operation with JSON serializer
func BenchmarkJSON_Get(b *testing.B) {
	config := cache.StoreConfig{
		Driver: "redis",
		Options: map[string]interface{}{
			"host":       "localhost",
			"port":       6379,
			"database":   15,
			"serializer": "json",
		},
	}

	driver, err := NewDriver(config)
	if err != nil {
		b.Skipf("Skipping benchmark: Redis not available: %v", err)
	}
	defer driver.Close()

	redisDriver := driver.(*Driver)
	ctx := context.Background()

	type User struct {
		ID    int
		Name  string
		Email string
	}

	user := User{ID: 1, Name: "John Doe", Email: "john@example.com"}
	redisDriver.Put(ctx, "bench:user", user, 1*time.Minute)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = redisDriver.Get(ctx, "bench:user")
	}
}

// BenchmarkMsgpack_Get benchmarks Get operation with msgpack serializer
func BenchmarkMsgpack_Get(b *testing.B) {
	config := cache.StoreConfig{
		Driver: "redis",
		Options: map[string]interface{}{
			"host":       "localhost",
			"port":       6379,
			"database":   15,
			"serializer": "msgpack",
		},
	}

	driver, err := NewDriver(config)
	if err != nil {
		b.Skipf("Skipping benchmark: Redis not available: %v", err)
	}
	defer driver.Close()

	redisDriver := driver.(*Driver)
	ctx := context.Background()

	type User struct {
		ID    int
		Name  string
		Email string
	}

	user := User{ID: 1, Name: "John Doe", Email: "john@example.com"}
	redisDriver.Put(ctx, "bench:user", user, 1*time.Minute)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = redisDriver.Get(ctx, "bench:user")
	}
}

// BenchmarkPutMultiple benchmarks batch Put operations
func BenchmarkPutMultiple(b *testing.B) {
	config := cache.StoreConfig{
		Driver: "redis",
		Options: map[string]interface{}{
			"host":       "localhost",
			"port":       6379,
			"database":   15,
			"serializer": "msgpack",
		},
	}

	driver, err := NewDriver(config)
	if err != nil {
		b.Skipf("Skipping benchmark: Redis not available: %v", err)
	}
	defer driver.Close()

	redisDriver := driver.(*Driver)
	ctx := context.Background()

	type User struct {
		ID   int
		Name string
	}

	items := map[string]interface{}{
		"user:1": User{ID: 1, Name: "User 1"},
		"user:2": User{ID: 2, Name: "User 2"},
		"user:3": User{ID: 3, Name: "User 3"},
		"user:4": User{ID: 4, Name: "User 4"},
		"user:5": User{ID: 5, Name: "User 5"},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = redisDriver.PutMultiple(ctx, items, 1*time.Minute)
	}
}

// BenchmarkGetMultiple benchmarks batch Get operations
func BenchmarkGetMultiple(b *testing.B) {
	config := cache.StoreConfig{
		Driver: "redis",
		Options: map[string]interface{}{
			"host":       "localhost",
			"port":       6379,
			"database":   15,
			"serializer": "msgpack",
		},
	}

	driver, err := NewDriver(config)
	if err != nil {
		b.Skipf("Skipping benchmark: Redis not available: %v", err)
	}
	defer driver.Close()

	redisDriver := driver.(*Driver)
	ctx := context.Background()

	type User struct {
		ID   int
		Name string
	}

	// Setup data
	items := map[string]interface{}{
		"user:1": User{ID: 1, Name: "User 1"},
		"user:2": User{ID: 2, Name: "User 2"},
		"user:3": User{ID: 3, Name: "User 3"},
		"user:4": User{ID: 4, Name: "User 4"},
		"user:5": User{ID: 5, Name: "User 5"},
	}
	redisDriver.PutMultiple(ctx, items, 1*time.Minute)

	keys := []string{"user:1", "user:2", "user:3", "user:4", "user:5"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = redisDriver.GetMultiple(ctx, keys)
	}
}

// BenchmarkTaggedCache_Put benchmarks tagged cache Put
func BenchmarkTaggedCache_Put(b *testing.B) {
	config := cache.StoreConfig{
		Driver: "redis",
		Options: map[string]interface{}{
			"host":       "localhost",
			"port":       6379,
			"database":   15,
			"serializer": "msgpack",
		},
	}

	driver, err := NewDriver(config)
	if err != nil {
		b.Skipf("Skipping benchmark: Redis not available: %v", err)
	}
	defer driver.Close()

	redisDriver := driver.(*Driver)
	ctx := context.Background()

	type Product struct {
		ID    int
		Name  string
		Price float64
	}

	product := Product{ID: 1, Name: "Widget", Price: 19.99}
	tagged := redisDriver.Tags("products", "active")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = tagged.Put(ctx, "product:1", product, 1*time.Minute)
	}
}

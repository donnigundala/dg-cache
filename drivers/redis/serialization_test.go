package redis

import (
	"context"
	"testing"
	"time"

	cache "github.com/donnigundala/dg-cache"
)

// TestSerialization_SimpleTypes tests serialization of primitive types
func TestSerialization_SimpleTypes(t *testing.T) {
	driver, cleanup := setupTestDriver(t)
	defer cleanup()

	ctx := context.Background()

	tests := []struct {
		name  string
		key   string
		value interface{}
	}{
		{"string", "test:string", "hello world"},
		{"int", "test:int", 42},
		{"int64", "test:int64", int64(9223372036854775807)},
		{"float64", "test:float64", 3.14159},
		{"bool", "test:bool", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Put
			err := driver.Put(ctx, tt.key, tt.value, 1*time.Minute)
			if err != nil {
				t.Fatalf("Put failed: %v", err)
			}

			// Get
			val, err := driver.Get(ctx, tt.key)
			if err != nil {
				t.Fatalf("Get failed: %v", err)
			}

			// Verify (note: JSON unmarshals numbers as float64)
			if tt.name == "int" || tt.name == "int64" {
				if f, ok := val.(float64); !ok {
					t.Errorf("Expected float64, got %T", val)
				} else if tt.name == "int" && int(f) != tt.value.(int) {
					t.Errorf("Value mismatch: expected %v, got %v", tt.value, int(f))
				}
			} else if val != tt.value {
				t.Errorf("Value mismatch: expected %v, got %v", tt.value, val)
			}
		})
	}
}

// TestSerialization_Struct tests struct serialization
func TestSerialization_Struct(t *testing.T) {
	driver, cleanup := setupTestDriver(t)
	defer cleanup()

	ctx := context.Background()

	type User struct {
		ID    int
		Name  string
		Email string
	}

	user := User{
		ID:    1,
		Name:  "John Doe",
		Email: "john@example.com",
	}

	// Put
	err := driver.Put(ctx, "test:user", user, 1*time.Minute)
	if err != nil {
		t.Fatalf("Put failed: %v", err)
	}

	// Get
	val, err := driver.Get(ctx, "test:user")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	// Verify - JSON unmarshals to map[string]interface{}
	if m, ok := val.(map[string]interface{}); ok {
		if m["Name"] != user.Name {
			t.Errorf("Name mismatch: expected %s, got %v", user.Name, m["Name"])
		}
		if m["Email"] != user.Email {
			t.Errorf("Email mismatch: expected %s, got %v", user.Email, m["Email"])
		}
	} else {
		t.Errorf("Expected map[string]interface{}, got %T", val)
	}
}

// TestSerialization_Slice tests slice serialization
func TestSerialization_Slice(t *testing.T) {
	driver, cleanup := setupTestDriver(t)
	defer cleanup()

	ctx := context.Background()

	items := []int{1, 2, 3, 4, 5}

	// Put
	err := driver.Put(ctx, "test:slice", items, 1*time.Minute)
	if err != nil {
		t.Fatalf("Put failed: %v", err)
	}

	// Get
	val, err := driver.Get(ctx, "test:slice")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	// Verify - JSON unmarshals to []interface{}
	if slice, ok := val.([]interface{}); ok {
		if len(slice) != len(items) {
			t.Errorf("Length mismatch: expected %d, got %d", len(items), len(slice))
		}
	} else {
		t.Errorf("Expected []interface{}, got %T", val)
	}
}

// TestSerialization_Map tests map serialization
func TestSerialization_Map(t *testing.T) {
	driver, cleanup := setupTestDriver(t)
	defer cleanup()

	ctx := context.Background()

	data := map[string]interface{}{
		"name":  "John",
		"age":   30,
		"admin": true,
	}

	// Put
	err := driver.Put(ctx, "test:map", data, 1*time.Minute)
	if err != nil {
		t.Fatalf("Put failed: %v", err)
	}

	// Get
	val, err := driver.Get(ctx, "test:map")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	// Verify
	if m, ok := val.(map[string]interface{}); ok {
		if m["name"] != "John" {
			t.Errorf("Name mismatch: expected John, got %v", m["name"])
		}
	} else {
		t.Errorf("Expected map[string]interface{}, got %T", val)
	}
}

// TestSerialization_PutMultiple tests batch serialization
func TestSerialization_PutMultiple(t *testing.T) {
	driver, cleanup := setupTestDriver(t)
	defer cleanup()

	ctx := context.Background()

	type User struct {
		ID   int
		Name string
	}

	items := map[string]interface{}{
		"user:1": User{ID: 1, Name: "John"},
		"user:2": User{ID: 2, Name: "Jane"},
		"user:3": "simple string",
	}

	// PutMultiple
	err := driver.PutMultiple(ctx, items, 1*time.Minute)
	if err != nil {
		t.Fatalf("PutMultiple failed: %v", err)
	}

	// GetMultiple
	keys := []string{"user:1", "user:2", "user:3"}
	results, err := driver.GetMultiple(ctx, keys)
	if err != nil {
		t.Fatalf("GetMultiple failed: %v", err)
	}

	// Verify
	if len(results) != 3 {
		t.Errorf("Expected 3 results, got %d", len(results))
	}

	// Verify string value
	if results["user:3"] != "simple string" {
		t.Errorf("String value mismatch: expected 'simple string', got %v", results["user:3"])
	}
}

// TestSerialization_BackwardCompatibility tests that raw strings still work
func TestSerialization_BackwardCompatibility(t *testing.T) {
	driver, cleanup := setupTestDriver(t)
	defer cleanup()

	ctx := context.Background()

	// Manually set a raw string value (simulating old cache)
	err := driver.client.Set(ctx, driver.prefixKey("test:raw"), "raw string value", 1*time.Minute).Err()
	if err != nil {
		t.Fatalf("Failed to set raw value: %v", err)
	}

	// Get should return it as a string
	val, err := driver.Get(ctx, "test:raw")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	if val != "raw string value" {
		t.Errorf("Expected 'raw string value', got %v", val)
	}
}

// TestTaggedCache_Serialization tests tagged cache with serialization
func TestTaggedCache_Serialization(t *testing.T) {
	driver, cleanup := setupTestDriver(t)
	defer cleanup()

	ctx := context.Background()

	type Product struct {
		ID    int
		Name  string
		Price float64
	}

	product := Product{
		ID:    1,
		Name:  "Widget",
		Price: 19.99,
	}

	// Create tagged cache
	tagged := driver.Tags("products", "active")

	// Put with tags
	err := tagged.Put(ctx, "product:1", product, 1*time.Minute)
	if err != nil {
		t.Fatalf("Tagged Put failed: %v", err)
	}

	// Get
	val, err := driver.Get(ctx, "product:1")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	// Verify
	if m, ok := val.(map[string]interface{}); ok {
		if m["Name"] != product.Name {
			t.Errorf("Name mismatch: expected %s, got %v", product.Name, m["Name"])
		}
	} else {
		t.Errorf("Expected map[string]interface{}, got %T", val)
	}
}

// TestTaggedCache_PutMultiple tests tagged batch operations
func TestTaggedCache_PutMultiple(t *testing.T) {
	driver, cleanup := setupTestDriver(t)
	defer cleanup()

	ctx := context.Background()

	type Item struct {
		ID   int
		Name string
	}

	items := map[string]interface{}{
		"item:1": Item{ID: 1, Name: "Item 1"},
		"item:2": Item{ID: 2, Name: "Item 2"},
	}

	// Create tagged cache
	tagged := driver.Tags("items")

	// PutMultiple with tags
	err := tagged.PutMultiple(ctx, items, 1*time.Minute)
	if err != nil {
		t.Fatalf("Tagged PutMultiple failed: %v", err)
	}

	// Verify items exist
	val, err := driver.Get(ctx, "item:1")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	if m, ok := val.(map[string]interface{}); ok {
		if m["Name"] != "Item 1" {
			t.Errorf("Name mismatch: expected 'Item 1', got %v", m["Name"])
		}
	}
}

// Helper function to setup test driver
func setupTestDriver(t *testing.T) (*Driver, func()) {
	config := cache.StoreConfig{
		Driver: "redis",
		Options: map[string]interface{}{
			"host":       "localhost",
			"port":       6379,
			"database":   15, // Use test database
			"serializer": "json",
		},
	}

	driver, err := NewDriver(config)
	if err != nil {
		t.Skipf("Skipping test: Redis not available: %v", err)
	}

	redisDriver := driver.(*Driver)

	// Cleanup function
	cleanup := func() {
		ctx := context.Background()
		redisDriver.Flush(ctx)
		redisDriver.Close()
	}

	return redisDriver, cleanup
}

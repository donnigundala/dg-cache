package cache

import (
	"context"
	"testing"
)

func TestManager_GetAs(t *testing.T) {
	manager, _ := NewManager(DefaultConfig())

	type User struct {
		ID   int
		Name string
	}

	ctx := context.Background()
	_ = User{ID: 1, Name: "John"} // Create but don't use (no store to test with)

	// This test would need a real store implementation
	// For now, just verify the method exists and compiles
	var result User
	_ = manager.GetAs(ctx, "user:1", &result)
}

func TestManager_GetString(t *testing.T) {
	manager, _ := NewManager(DefaultConfig())
	ctx := context.Background()

	// Verify method exists
	_, _ = manager.GetString(ctx, "key")
}

func TestManager_GetInt(t *testing.T) {
	manager, _ := NewManager(DefaultConfig())
	ctx := context.Background()

	// Verify method exists
	_, _ = manager.GetInt(ctx, "key")
}

func TestManager_GetInt64(t *testing.T) {
	manager, _ := NewManager(DefaultConfig())
	ctx := context.Background()

	// Verify method exists
	_, _ = manager.GetInt64(ctx, "key")
}

func TestManager_GetFloat64(t *testing.T) {
	manager, _ := NewManager(DefaultConfig())
	ctx := context.Background()

	// Verify method exists
	_, _ = manager.GetFloat64(ctx, "key")
}

func TestManager_GetBool(t *testing.T) {
	manager, _ := NewManager(DefaultConfig())
	ctx := context.Background()

	// Verify method exists
	_, _ = manager.GetBool(ctx, "key")
}

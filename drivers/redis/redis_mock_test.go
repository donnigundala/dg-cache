package redis_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestRedis_Errors(t *testing.T) {
	d, s := createDriver(t)
	defer s.Close()
	defer d.Close()

	ctx := context.Background()

	// Simulate Redis failure
	s.SetError("redis is down")

	t.Run("Get returns error on failure", func(t *testing.T) {
		val, err := d.Get(ctx, "key")
		assert.Error(t, err)
		assert.Nil(t, val)
		assert.Contains(t, err.Error(), "redis is down")
	})

	t.Run("Put returns error on failure", func(t *testing.T) {
		err := d.Put(ctx, "key", "value", 1*time.Minute)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "redis is down")
	})

	t.Run("Forget returns error on failure", func(t *testing.T) {
		err := d.Forget(ctx, "key")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "redis is down")
	})

	t.Run("Flush returns error on failure", func(t *testing.T) {
		err := d.Flush(ctx)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "redis is down")
	})

	t.Run("Has returns error on failure", func(t *testing.T) {
		exists, err := d.Has(ctx, "key")
		assert.Error(t, err)
		assert.False(t, exists)
		assert.Contains(t, err.Error(), "redis is down")
	})
}

func TestRedis_MultiErrors(t *testing.T) {
	d, s := createDriver(t)
	defer s.Close()
	defer d.Close()

	ctx := context.Background()
	s.SetError("redis is down")

	t.Run("GetMultiple returns error on failure", func(t *testing.T) {
		vals, err := d.GetMultiple(ctx, []string{"k1", "k2"})
		assert.Error(t, err)
		assert.Nil(t, vals)
		assert.Contains(t, err.Error(), "redis is down")
	})

	t.Run("PutMultiple returns error on failure", func(t *testing.T) {
		items := map[string]interface{}{"k1": "v1"}
		err := d.PutMultiple(ctx, items, 1*time.Minute)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "redis is down")
	})
}

package redis

import (
	"context"
	"time"

	cache "github.com/donnigundala/dg-cache"
	"github.com/donnigundala/dg-cache/serializer"
	"github.com/redis/go-redis/v9"
)

// Driver is a Redis cache driver.
type Driver struct {
	client     *redis.Client
	prefix     string
	serializer serializer.Serializer
}

// NewDriver creates a new Redis cache driver.
func NewDriver(config cache.StoreConfig) (cache.Driver, error) {
	// Parse options into Redis config
	redisConfig := DefaultConfig()

	if val, ok := config.Options["host"].(string); ok {
		redisConfig.Host = val
	}
	if val, ok := config.Options["port"].(int); ok {
		redisConfig.Port = val
	}
	if val, ok := config.Options["password"].(string); ok {
		redisConfig.Password = val
	}
	if val, ok := config.Options["database"].(int); ok {
		redisConfig.Database = val
	}
	if val, ok := config.Options["pool_size"].(int); ok {
		redisConfig.PoolSize = val
	}

	client, err := NewClient(redisConfig)
	if err != nil {
		return nil, err
	}

	// Initialize serializer (default to JSON)
	var ser serializer.Serializer = serializer.NewJSONSerializer()
	if val, ok := config.Options["serializer"].(string); ok {
		switch val {
		case "msgpack":
			ser = serializer.NewMsgpackSerializer()
		case "json":
			ser = serializer.NewJSONSerializer()
		}
	}

	return &Driver{
		client:     client,
		prefix:     config.Prefix,
		serializer: ser,
	}, nil
}

// NewDriverWithClient creates a new Redis cache driver with an existing client.
func NewDriverWithClient(client *redis.Client, prefix string) *Driver {
	return &Driver{
		client:     client,
		prefix:     prefix,
		serializer: serializer.NewJSONSerializer(), // Default to JSON
	}
}

// prefixKey adds the prefix to the key.
func (d *Driver) prefixKey(key string) string {
	if d.prefix == "" {
		return key
	}
	return d.prefix + ":" + key
}

// Get retrieves a value from the cache.
func (d *Driver) Get(ctx context.Context, key string) (interface{}, error) {
	data, err := d.client.Get(ctx, d.prefixKey(key)).Bytes()
	if err == redis.Nil {
		return nil, cache.ErrKeyNotFound
	}
	if err != nil {
		return nil, err
	}

	// Try to deserialize
	var result interface{}
	if err := d.serializer.Unmarshal(data, &result); err != nil {
		// Fallback: return as string for backward compatibility
		return string(data), nil
	}

	return result, nil
}

// GetMultiple retrieves multiple values from the cache.
func (d *Driver) GetMultiple(ctx context.Context, keys []string) (map[string]interface{}, error) {
	prefixedKeys := make([]string, len(keys))
	for i, key := range keys {
		prefixedKeys[i] = d.prefixKey(key)
	}

	vals, err := d.client.MGet(ctx, prefixedKeys...).Result()
	if err != nil {
		return nil, err
	}

	result := make(map[string]interface{})
	for i, val := range vals {
		if val != nil {
			// Convert to bytes for deserialization
			var data []byte
			switch v := val.(type) {
			case string:
				data = []byte(v)
			case []byte:
				data = v
			default:
				continue // Skip if not string or bytes
			}

			// Try to deserialize
			var value interface{}
			if err := d.serializer.Unmarshal(data, &value); err != nil {
				// Fallback: use as string
				result[keys[i]] = string(data)
			} else {
				result[keys[i]] = value
			}
		}
	}

	return result, nil
}

// Put stores a value in the cache with the given TTL.
func (d *Driver) Put(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	data, err := d.serializer.Marshal(value)
	if err != nil {
		return err
	}
	return d.client.Set(ctx, d.prefixKey(key), data, ttl).Err()
}

// PutMultiple stores multiple values in the cache.
func (d *Driver) PutMultiple(ctx context.Context, items map[string]interface{}, ttl time.Duration) error {
	pipe := d.client.Pipeline()
	for key, value := range items {
		// Serialize each value
		data, err := d.serializer.Marshal(value)
		if err != nil {
			return err
		}
		pipe.Set(ctx, d.prefixKey(key), data, ttl)
	}
	_, err := pipe.Exec(ctx)
	return err
}

// Increment increments the value of a key.
func (d *Driver) Increment(ctx context.Context, key string, value int64) (int64, error) {
	return d.client.IncrBy(ctx, d.prefixKey(key), value).Result()
}

// Decrement decrements the value of a key.
func (d *Driver) Decrement(ctx context.Context, key string, value int64) (int64, error) {
	return d.client.DecrBy(ctx, d.prefixKey(key), value).Result()
}

// Forever stores a value in the cache indefinitely.
func (d *Driver) Forever(ctx context.Context, key string, value interface{}) error {
	return d.Put(ctx, key, value, 0)
}

// Forget removes a value from the cache.
func (d *Driver) Forget(ctx context.Context, key string) error {
	return d.client.Del(ctx, d.prefixKey(key)).Err()
}

// Flush removes all items from the cache.
func (d *Driver) Flush(ctx context.Context) error {
	return d.client.FlushDB(ctx).Err()
}

// Has checks if a key exists in the cache.
func (d *Driver) Has(ctx context.Context, key string) (bool, error) {
	n, err := d.client.Exists(ctx, d.prefixKey(key)).Result()
	if err != nil {
		return false, err
	}
	return n > 0, nil
}

// Missing checks if a key does not exist in the cache.
func (d *Driver) Missing(ctx context.Context, key string) (bool, error) {
	has, err := d.Has(ctx, key)
	return !has, err
}

// GetPrefix returns the cache key prefix.
func (d *Driver) GetPrefix() string {
	return d.prefix
}

// SetPrefix sets the cache key prefix.
func (d *Driver) SetPrefix(prefix string) {
	d.prefix = prefix
}

// Name returns the driver name.
func (d *Driver) Name() string {
	return "redis"
}

// Close closes the driver and releases resources.
func (d *Driver) Close() error {
	return d.client.Close()
}

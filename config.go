package cache

import (
	"time"

	"github.com/mitchellh/mapstructure"
)

// Config represents the cache configuration.
type Config struct {
	// DefaultStore is the name of the default cache store to use.
	DefaultStore string `mapstructure:"default_store"`

	// Prefix is the global cache key prefix.
	// This is prepended to all cache keys.
	Prefix string `mapstructure:"prefix"`

	// Stores contains the configuration for each cache store.
	Stores map[string]StoreConfig `mapstructure:"stores"`
}

// StoreConfig represents the configuration for a single cache store.
type StoreConfig struct {
	// Driver is the cache driver name (e.g., "redis", "memory").
	Driver string `mapstructure:"driver"`

	// Connection is the connection name to use (for drivers that support multiple connections).
	Connection string `mapstructure:"connection"`

	// Prefix is the store-specific cache key prefix.
	// This overrides the global prefix for this store.
	Prefix string `mapstructure:"prefix"`

	// Options contains driver-specific configuration options.
	Options map[string]interface{} `mapstructure:"options"`
}

// Decode decodes the store options into the target struct.
func (c StoreConfig) Decode(target interface{}) error {
	decoder, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		Metadata: nil,
		Result:   target,
		TagName:  "mapstructure",
	})
	if err != nil {
		return err
	}
	return decoder.Decode(c.Options)
}

// DefaultConfig returns a default cache configuration.
func DefaultConfig() Config {
	return Config{
		DefaultStore: "memory",
		Prefix:       "cache",
		Stores: map[string]StoreConfig{
			"memory": {
				Driver: "memory",
			},
		},
	}
}

// WithDefaultStore sets the default store name.
func (c Config) WithDefaultStore(name string) Config {
	c.DefaultStore = name
	return c
}

// WithPrefix sets the global cache key prefix.
func (c Config) WithPrefix(prefix string) Config {
	c.Prefix = prefix
	return c
}

// WithStore adds a store configuration.
func (c Config) WithStore(name string, config StoreConfig) Config {
	if c.Stores == nil {
		c.Stores = make(map[string]StoreConfig)
	}
	c.Stores[name] = config
	return c
}

// Validate validates the cache configuration.
func (c Config) Validate() error {
	if c.DefaultStore == "" {
		return ErrInvalidConfig("default store is required")
	}

	if c.Stores == nil || len(c.Stores) == 0 {
		return ErrInvalidConfig("at least one store must be configured")
	}

	if _, ok := c.Stores[c.DefaultStore]; !ok {
		return ErrInvalidConfig("default store '%s' is not configured", c.DefaultStore)
	}

	for name, store := range c.Stores {
		if store.Driver == "" {
			return ErrInvalidConfig("driver is required for store '%s'", name)
		}
	}

	return nil
}

// Item represents a cache item with metadata.
type Item struct {
	// Key is the cache key.
	Key string

	// Value is the cached value.
	Value interface{}

	// ExpiresAt is the expiration time.
	// Zero value means the item never expires.
	ExpiresAt time.Time

	// Tags are the tags associated with this item.
	Tags []string
}

// IsExpired checks if the item has expired.
func (i Item) IsExpired() bool {
	if i.ExpiresAt.IsZero() {
		return false
	}
	return time.Now().After(i.ExpiresAt)
}

// TTL returns the time until expiration.
// Returns 0 if the item has expired or never expires.
func (i Item) TTL() time.Duration {
	if i.ExpiresAt.IsZero() {
		return 0
	}
	ttl := time.Until(i.ExpiresAt)
	if ttl < 0 {
		return 0
	}
	return ttl
}

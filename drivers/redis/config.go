package redis

import (
	"time"
)

// Config represents the Redis configuration.
type Config struct {
	// Host is the Redis server host.
	Host string

	// Port is the Redis server port.
	Port int

	// Password is the Redis server password.
	Password string

	// Database is the Redis database number.
	Database int

	// Prefix is the cache key prefix.
	Prefix string

	// PoolSize is the maximum number of socket connections.
	PoolSize int

	// MinIdleConns is the minimum number of idle connections.
	MinIdleConns int

	// MaxRetries is the maximum number of retries before giving up.
	MaxRetries int

	// Timeout is the dial timeout.
	Timeout time.Duration

	// MinRetryBackoff is the minimum backoff between retries.
	MinRetryBackoff time.Duration

	// MaxRetryBackoff is the maximum backoff between retries.
	MaxRetryBackoff time.Duration
}

// DefaultConfig returns a default Redis configuration.
func DefaultConfig() Config {
	return Config{
		Host:            "localhost",
		Port:            6379,
		Database:        0,
		PoolSize:        10,
		MinIdleConns:    2,
		MaxRetries:      3,
		Timeout:         5 * time.Second,
		MinRetryBackoff: 8 * time.Millisecond,
		MaxRetryBackoff: 512 * time.Millisecond,
	}
}

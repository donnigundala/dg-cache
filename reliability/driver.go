package reliability

import (
	"context"
	"time"

	cache "github.com/donnigundala/dg-cache"
)

// CircuitBreakerDriver wraps a cache driver with a circuit breaker.
type CircuitBreakerDriver struct {
	cache.Driver
	breaker Breaker
}

// NewCircuitBreakerDriver creates a new CircuitBreakerDriver.
func NewCircuitBreakerDriver(driver cache.Driver, breaker Breaker) *CircuitBreakerDriver {
	return &CircuitBreakerDriver{
		Driver:  driver,
		breaker: breaker,
	}
}

func (d *CircuitBreakerDriver) Get(ctx context.Context, key string) (interface{}, error) {
	if !d.breaker.Allow() {
		return nil, ErrCircuitOpen
	}
	val, err := d.Driver.Get(ctx, key)
	d.report(err)
	return val, err
}

func (d *CircuitBreakerDriver) Put(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	if !d.breaker.Allow() {
		return ErrCircuitOpen
	}
	err := d.Driver.Put(ctx, key, value, ttl)
	d.report(err)
	return err
}

func (d *CircuitBreakerDriver) Forget(ctx context.Context, key string) error {
	if !d.breaker.Allow() {
		return ErrCircuitOpen
	}
	err := d.Driver.Forget(ctx, key)
	d.report(err)
	return err
}

func (d *CircuitBreakerDriver) Flush(ctx context.Context) error {
	if !d.breaker.Allow() {
		return ErrCircuitOpen
	}
	err := d.Driver.Flush(ctx)
	d.report(err)
	return err
}

// report updates the breaker state based on the error.
func (d *CircuitBreakerDriver) report(err error) {
	if err != nil && err != cache.ErrKeyNotFound {
		d.breaker.Failure()
	} else {
		d.breaker.Success()
	}
}

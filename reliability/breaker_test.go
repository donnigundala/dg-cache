package reliability

import (
	"context"
	"errors"
	"testing"
	"time"

	cache "github.com/donnigundala/dg-cache"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockDriver is a mock cache driver.
type MockDriver struct {
	mock.Mock
}

func (m *MockDriver) Get(ctx context.Context, key string) (interface{}, error) {
	args := m.Called(ctx, key)
	return args.Get(0), args.Error(1)
}

func (m *MockDriver) Put(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	return m.Called(ctx, key, value, ttl).Error(0)
}

// Add other required methods to satisfy Driver interface...
func (m *MockDriver) Has(ctx context.Context, key string) (bool, error) {
	return false, nil
}
func (m *MockDriver) Forget(ctx context.Context, key string) error {
	return nil
}
func (m *MockDriver) Flush(ctx context.Context) error {
	return nil
}
func (m *MockDriver) GetMultiple(ctx context.Context, keys []string) (map[string]interface{}, error) {
	return nil, nil
}
func (m *MockDriver) PutMultiple(ctx context.Context, items map[string]interface{}, ttl time.Duration) error {
	return nil
}
func (m *MockDriver) Increment(ctx context.Context, key string, value int64) (int64, error) {
	return 0, nil
}
func (m *MockDriver) Decrement(ctx context.Context, key string, value int64) (int64, error) {
	return 0, nil
}
func (m *MockDriver) Forever(ctx context.Context, key string, value interface{}) error {
	return nil
}
func (m *MockDriver) ForgetMultiple(ctx context.Context, keys []string) error {
	return nil
}
func (m *MockDriver) Missing(ctx context.Context, key string) (bool, error) {
	return false, nil
}
func (m *MockDriver) GetPrefix() string {
	return ""
}
func (m *MockDriver) SetPrefix(prefix string) {}
func (m *MockDriver) Name() string {
	return "mock"
}
func (m *MockDriver) Close() error {
	return nil
}
func (m *MockDriver) Stats() cache.Stats {
	return cache.Stats{}
}

func TestThresholdBreaker(t *testing.T) {
	breaker := NewThresholdBreaker(3, 100*time.Millisecond)

	// Initially closed
	assert.True(t, breaker.Allow())

	// Fail 2 times (should stay closed)
	breaker.Failure()
	breaker.Failure()
	assert.True(t, breaker.Allow())

	// Fail 3rd time (should trip)
	breaker.Failure()
	assert.False(t, breaker.Allow())

	// Wait for timeout (half-open)
	time.Sleep(150 * time.Millisecond)
	assert.True(t, breaker.Allow())

	// Success (should close)
	breaker.Success()
	assert.True(t, breaker.Allow())
}

func TestCircuitBreakerDriver(t *testing.T) {
	mockDriver := new(MockDriver)
	breaker := NewThresholdBreaker(1, 1*time.Second)
	driver := NewCircuitBreakerDriver(mockDriver, breaker)

	ctx := context.Background()

	// 1. Success case
	mockDriver.On("Get", ctx, "key1").Return("value", nil)
	val, err := driver.Get(ctx, "key1")
	assert.NoError(t, err)
	assert.Equal(t, "value", val)

	// 2. Failure case (error triggers failure)
	mockDriver.On("Get", ctx, "key2").Return(nil, errors.New("db error"))
	_, err = driver.Get(ctx, "key2")
	assert.Error(t, err)

	// 3. Circuit should now be open
	_, err = driver.Get(ctx, "key3")
	assert.Equal(t, ErrCircuitOpen, err)
}

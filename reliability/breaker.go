package reliability

import (
	"errors"
	"sync"
	"time"
)

var (
	// ErrCircuitOpen is returned when the circuit breaker is open.
	ErrCircuitOpen = errors.New("circuit breaker is open")
)

// Breaker defines the interface for a circuit breaker.
type Breaker interface {
	// Allow checks if the request is allowed to proceed.
	Allow() bool

	// Success reports a successful execution.
	Success()

	// Failure reports a failed execution.
	Failure()
}

// State represents the state of the circuit breaker.
type State int

const (
	StateClosed State = iota
	StateOpen
	StateHalfOpen
)

// ThresholdBreaker implements a simple failure threshold circuit breaker.
type ThresholdBreaker struct {
	mu sync.Mutex

	state            State
	failures         int
	failureThreshold int
	resetTimeout     time.Duration
	lastFailureTime  time.Time
}

// NewThresholdBreaker creates a new ThresholdBreaker.
func NewThresholdBreaker(threshold int, timeout time.Duration) *ThresholdBreaker {
	return &ThresholdBreaker{
		state:            StateClosed,
		failureThreshold: threshold,
		resetTimeout:     timeout,
	}
}

// Allow checks if the request is allowed.
func (b *ThresholdBreaker) Allow() bool {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.state == StateOpen {
		if time.Since(b.lastFailureTime) > b.resetTimeout {
			b.state = StateHalfOpen
			return true
		}
		return false
	}

	return true
}

// Success reports a success.
func (b *ThresholdBreaker) Success() {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.state == StateHalfOpen {
		b.state = StateClosed
		b.failures = 0
	} else if b.state == StateClosed {
		b.failures = 0
	}
}

// Failure reports a failure.
func (b *ThresholdBreaker) Failure() {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.state == StateClosed {
		b.failures++
		if b.failures >= b.failureThreshold {
			b.state = StateOpen
			b.lastFailureTime = time.Now()
		}
	} else if b.state == StateHalfOpen {
		b.state = StateOpen
		b.lastFailureTime = time.Now()
	}
}

package service

import (
	"fmt"
	"sync"
	"time"
)

// CircuitState represents the state of a circuit breaker.
type CircuitState int

const (
	CircuitClosed   CircuitState = iota // normal operation
	CircuitOpen                         // failing, reject requests
	CircuitHalfOpen                     // testing recovery — single probe allowed
)

// String returns a human-readable state name.
func (s CircuitState) String() string {
	switch s {
	case CircuitClosed:
		return "closed"
	case CircuitOpen:
		return "open"
	case CircuitHalfOpen:
		return "half-open"
	default:
		return fmt.Sprintf("unknown(%d)", int(s))
	}
}

// CircuitBreaker implements the circuit breaker pattern.
// It is safe for concurrent use.
type CircuitBreaker struct {
	mu           sync.Mutex
	state        CircuitState
	failures     int
	lastFailure  time.Time
	probing      bool // true when a half-open probe is in flight
	maxFailures  int
	resetTimeout time.Duration
	windowSize   time.Duration
}

// NewCircuitBreaker creates a new circuit breaker.
func NewCircuitBreaker(maxFailures int, resetTimeout, windowSize time.Duration) *CircuitBreaker {
	return &CircuitBreaker{
		state:        CircuitClosed,
		maxFailures:  maxFailures,
		resetTimeout: resetTimeout,
		windowSize:   windowSize,
	}
}

// Allow checks if a request should be allowed through.
func (cb *CircuitBreaker) Allow() bool {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	switch cb.state {
	case CircuitClosed:
		return true
	case CircuitOpen:
		if time.Since(cb.lastFailure) > cb.resetTimeout {
			cb.state = CircuitHalfOpen
			cb.probing = true
			return true
		}
		return false
	case CircuitHalfOpen:
		// Only allow one probe at a time in half-open state
		if cb.probing {
			return false
		}
		cb.probing = true
		return true
	}
	return false
}

// RecordSuccess records a successful request.
func (cb *CircuitBreaker) RecordSuccess() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	switch cb.state {
	case CircuitHalfOpen:
		// Single successful probe closes the circuit
		cb.state = CircuitClosed
		cb.failures = 0
		cb.probing = false
	case CircuitClosed:
		// Decrement failure count on success so intermittent errors don't
		// accumulate indefinitely within the window.
		if cb.failures > 0 {
			cb.failures--
		}
	}
}

// RecordFailure records a failed request.
func (cb *CircuitBreaker) RecordFailure() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	now := time.Now()

	switch cb.state {
	case CircuitClosed:
		// Reset failure count if outside the window
		if !cb.lastFailure.IsZero() && now.Sub(cb.lastFailure) > cb.windowSize {
			cb.failures = 0
		}
		cb.failures++
		cb.lastFailure = now
		if cb.failures >= cb.maxFailures {
			cb.state = CircuitOpen
		}
	case CircuitHalfOpen:
		cb.state = CircuitOpen
		cb.lastFailure = now
		cb.probing = false
	}
}

// State returns the current circuit state.
func (cb *CircuitBreaker) State() CircuitState {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	return cb.state
}

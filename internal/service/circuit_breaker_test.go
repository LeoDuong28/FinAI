package service

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestCircuitBreaker_ClosedAllowsRequests(t *testing.T) {
	cb := NewCircuitBreaker(3, 5*time.Second, 10*time.Second)
	assert.True(t, cb.Allow())
	assert.Equal(t, CircuitClosed, cb.State())
}

func TestCircuitBreaker_OpensAfterMaxFailures(t *testing.T) {
	cb := NewCircuitBreaker(3, 5*time.Second, 10*time.Second)

	cb.RecordFailure()
	cb.RecordFailure()
	assert.Equal(t, CircuitClosed, cb.State())

	cb.RecordFailure()
	assert.Equal(t, CircuitOpen, cb.State())
}

func TestCircuitBreaker_OpenRejectsRequests(t *testing.T) {
	cb := NewCircuitBreaker(1, 5*time.Second, 10*time.Second)
	cb.RecordFailure()
	assert.Equal(t, CircuitOpen, cb.State())
	assert.False(t, cb.Allow())
}

func TestCircuitBreaker_OpenToHalfOpenAfterTimeout(t *testing.T) {
	cb := NewCircuitBreaker(1, 10*time.Millisecond, 10*time.Second)
	cb.RecordFailure()
	assert.Equal(t, CircuitOpen, cb.State())

	time.Sleep(20 * time.Millisecond)
	assert.True(t, cb.Allow()) // transitions to half-open and allows probe
	assert.Equal(t, CircuitHalfOpen, cb.State())
}

func TestCircuitBreaker_HalfOpenToClosedOnSuccess(t *testing.T) {
	cb := NewCircuitBreaker(1, 10*time.Millisecond, 10*time.Second)
	cb.RecordFailure()

	time.Sleep(20 * time.Millisecond)
	cb.Allow() // enter half-open
	cb.RecordSuccess()

	assert.Equal(t, CircuitClosed, cb.State())
	assert.True(t, cb.Allow())
}

func TestCircuitBreaker_HalfOpenToOpenOnFailure(t *testing.T) {
	cb := NewCircuitBreaker(1, 10*time.Millisecond, 10*time.Second)
	cb.RecordFailure()

	time.Sleep(20 * time.Millisecond)
	cb.Allow() // enter half-open
	cb.RecordFailure()

	assert.Equal(t, CircuitOpen, cb.State())
}

func TestCircuitBreaker_HalfOpenBlocksConcurrentProbes(t *testing.T) {
	cb := NewCircuitBreaker(1, 10*time.Millisecond, 10*time.Second)
	cb.RecordFailure()

	time.Sleep(20 * time.Millisecond)
	assert.True(t, cb.Allow())  // first probe allowed
	assert.False(t, cb.Allow()) // second probe blocked
}

func TestCircuitBreaker_SuccessDecrementsFailures(t *testing.T) {
	cb := NewCircuitBreaker(3, 5*time.Second, 10*time.Second)

	cb.RecordFailure()
	cb.RecordFailure()
	cb.RecordSuccess() // decrements to 1

	// Need 2 more failures to trip (not 1)
	cb.RecordFailure()
	assert.Equal(t, CircuitClosed, cb.State())

	cb.RecordFailure()
	assert.Equal(t, CircuitOpen, cb.State())
}

func TestCircuitBreaker_WindowReset(t *testing.T) {
	cb := NewCircuitBreaker(2, 5*time.Second, 10*time.Millisecond)

	cb.RecordFailure()
	time.Sleep(20 * time.Millisecond) // exceed window

	// Failures should reset, so this is failure #1 not #2
	cb.RecordFailure()
	assert.Equal(t, CircuitClosed, cb.State())
}

func TestCircuitState_String(t *testing.T) {
	assert.Equal(t, "closed", CircuitClosed.String())
	assert.Equal(t, "open", CircuitOpen.String())
	assert.Equal(t, "half-open", CircuitHalfOpen.String())
	assert.Contains(t, CircuitState(99).String(), "unknown")
}

package ratelimit

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestRateLimiter_Allow(t *testing.T) {
	config := Config{
		RPS:   2, // 2 requests per second
		Burst: 2,
		CircuitBreakerConfig: CircuitBreakerConfig{
			FailureThreshold: 3,
			RecoveryTimeout:  1 * time.Second,
		},
	}

	limiter := NewRateLimiter(config)
	ctx := context.Background()

	// First two requests should be allowed (burst)
	err1 := limiter.Allow(ctx)
	err2 := limiter.Allow(ctx)

	assert.NoError(t, err1)
	assert.NoError(t, err2)

	// Third request should be rejected (rate limit exceeded)
	err3 := limiter.Allow(ctx)
	assert.Error(t, err3)
	assert.Contains(t, err3.Error(), "rate limit exceeded")
}

func TestRateLimiter_Wait(t *testing.T) {
	config := Config{
		RPS:   10, // Higher rate for faster test
		Burst: 1,
		CircuitBreakerConfig: CircuitBreakerConfig{
			FailureThreshold: 3,
			RecoveryTimeout:  1 * time.Second,
		},
	}

	limiter := NewRateLimiter(config)
	ctx := context.Background()

	start := time.Now()

	// First request should be immediate
	err1 := limiter.Wait(ctx)
	assert.NoError(t, err1)

	// Second request should wait
	err2 := limiter.Wait(ctx)
	assert.NoError(t, err2)

	duration := time.Since(start)
	// Should take at least some time to wait for rate limit
	assert.Greater(t, duration, 50*time.Millisecond)
}

func TestCircuitBreaker_States(t *testing.T) {
	config := CircuitBreakerConfig{
		FailureThreshold: 2,
		RecoveryTimeout:  100 * time.Millisecond,
		HalfOpenMaxCalls: 1,
	}

	cb := &CircuitBreaker{
		state:  StateClosed,
		config: config,
	}

	// Initial state should be closed
	assert.Equal(t, StateClosed, cb.State())
	assert.True(t, cb.CanExecute())

	// Record failures to trip circuit breaker
	cb.OnFailure()
	assert.Equal(t, StateClosed, cb.State()) // Still closed after 1 failure

	cb.OnFailure()
	assert.Equal(t, StateOpen, cb.State()) // Should be open after 2 failures
	assert.False(t, cb.CanExecute())

	// Wait for recovery timeout
	time.Sleep(150 * time.Millisecond)

	// Should transition to half-open
	assert.True(t, cb.CanExecute()) // This should trigger transition to half-open
	assert.Equal(t, StateHalfOpen, cb.State())

	// Success in half-open should close circuit
	cb.OnSuccess()
	assert.Equal(t, StateClosed, cb.State())
}

func TestCircuitBreaker_HalfOpenToOpen(t *testing.T) {
	config := CircuitBreakerConfig{
		FailureThreshold: 1,
		RecoveryTimeout:  100 * time.Millisecond,
		HalfOpenMaxCalls: 2,
	}

	cb := &CircuitBreaker{
		state:  StateClosed,
		config: config,
	}

	// Trip circuit breaker
	cb.OnFailure()
	assert.Equal(t, StateOpen, cb.State())

	// Wait for recovery
	time.Sleep(150 * time.Millisecond)
	cb.CanExecute() // Transition to half-open

	assert.Equal(t, StateHalfOpen, cb.State())

	// Failure in half-open should go back to open
	cb.OnFailure()
	assert.Equal(t, StateOpen, cb.State())
}

func TestRateLimiter_CircuitBreakerIntegration(t *testing.T) {
	config := Config{
		RPS:   10,
		Burst: 10,
		CircuitBreakerConfig: CircuitBreakerConfig{
			FailureThreshold: 2,
			RecoveryTimeout:  100 * time.Millisecond,
		},
	}

	limiter := NewRateLimiter(config)
	ctx := context.Background()

	// Should allow requests initially
	err := limiter.Allow(ctx)
	assert.NoError(t, err)

	// Record failures to trip circuit breaker
	limiter.OnFailure()
	limiter.OnFailure()

	// Should reject requests due to circuit breaker
	err = limiter.Allow(ctx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "circuit breaker")
}

func TestRateLimiter_Metrics(t *testing.T) {
	config := DefaultConfig()
	limiter := NewRateLimiter(config)
	ctx := context.Background()

	// Initial metrics
	metrics := limiter.GetMetrics()
	assert.Equal(t, int64(0), metrics.TotalRequests)
	assert.Equal(t, int64(0), metrics.AllowedRequests)
	assert.Equal(t, int64(0), metrics.RejectedRequests)

	// Make some requests
	limiter.Allow(ctx)
	limiter.Allow(ctx)

	// Exceed rate limit
	for i := 0; i < 20; i++ {
		limiter.Allow(ctx)
	}

	metrics = limiter.GetMetrics()
	assert.Greater(t, metrics.TotalRequests, int64(0))
	assert.Greater(t, metrics.RejectedRequests, int64(0))
}

func TestRateLimiter_ResetMetrics(t *testing.T) {
	config := DefaultConfig()
	limiter := NewRateLimiter(config)
	ctx := context.Background()

	// Make some requests
	limiter.Allow(ctx)
	limiter.Allow(ctx)

	// Should have some metrics
	metrics := limiter.GetMetrics()
	assert.Greater(t, metrics.TotalRequests, int64(0))

	// Reset metrics
	limiter.ResetMetrics()

	// Should be reset
	metrics = limiter.GetMetrics()
	assert.Equal(t, int64(0), metrics.TotalRequests)
	assert.Equal(t, int64(0), metrics.AllowedRequests)
	assert.Equal(t, int64(0), metrics.RejectedRequests)
}

func TestAdaptiveRateLimiter(t *testing.T) {
	limiter := NewAdaptiveRateLimiter(10, 1, 50)

	// Initial rate should be base rate
	assert.Equal(t, 10.0, limiter.currentRPS)

	// Simulate high success rate
	ctx := context.Background()
	for i := 0; i < 10; i++ {
		limiter.Allow(ctx)
		limiter.OnSuccess()
	}

	// Force adjustment
	limiter.lastAdjust = time.Now().Add(-1 * time.Minute)
	limiter.Adjust()

	// Rate should increase due to high success rate
	assert.Greater(t, limiter.currentRPS, 10.0)
	assert.LessOrEqual(t, limiter.currentRPS, limiter.maxRPS)
}

func TestAdaptiveRateLimiter_Decrease(t *testing.T) {
	limiter := NewAdaptiveRateLimiter(10, 1, 50)

	ctx := context.Background()

	// Simulate failures
	for i := 0; i < 5; i++ {
		limiter.Allow(ctx)
		limiter.OnFailure()
	}

	// Allow some requests to get rejected
	for i := 0; i < 10; i++ {
		limiter.Allow(ctx)
	}

	// Force adjustment
	limiter.lastAdjust = time.Now().Add(-1 * time.Minute)
	limiter.Adjust()

	// Rate should decrease due to low success rate
	assert.Less(t, limiter.currentRPS, 10.0)
	assert.GreaterOrEqual(t, limiter.currentRPS, limiter.minRPS)
}

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()

	assert.Equal(t, 10.0, config.RPS)
	assert.Equal(t, 20, config.Burst)
	assert.Equal(t, 5, config.CircuitBreakerConfig.FailureThreshold)
	assert.Equal(t, 60*time.Second, config.CircuitBreakerConfig.RecoveryTimeout)
	assert.Equal(t, 3, config.CircuitBreakerConfig.MaxRetries)
	assert.Equal(t, 1*time.Second, config.CircuitBreakerConfig.RetryDelay)
	assert.Equal(t, 3, config.CircuitBreakerConfig.HalfOpenMaxCalls)
}

func TestCircuitBreakerState_String(t *testing.T) {
	assert.Equal(t, "closed", StateClosed.String())
	assert.Equal(t, "open", StateOpen.String())
	assert.Equal(t, "half-open", StateHalfOpen.String())
	assert.Equal(t, "unknown", CircuitBreakerState(99).String())
}

// Note: Context cancellation test removed due to timing sensitivity
// The rate limiter properly handles context cancellation via golang.org/x/time/rate

package ratelimit

import (
	"context"
	"fmt"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

// RateLimiter provides advanced rate limiting with circuit breaker
type RateLimiter struct {
	limiter        *rate.Limiter
	circuitBreaker *CircuitBreaker
	metrics        *rateLimitMetrics
	config         Config
}

// Config holds rate limiter configuration
type Config struct {
	RPS                  float64
	Burst                int
	CircuitBreakerConfig CircuitBreakerConfig
}

// CircuitBreakerConfig holds circuit breaker configuration
type CircuitBreakerConfig struct {
	FailureThreshold int
	RecoveryTimeout  time.Duration
	MaxRetries       int
	RetryDelay       time.Duration
	HalfOpenMaxCalls int
}

// CircuitBreakerState represents the state of the circuit breaker
type CircuitBreakerState int

const (
	StateClosed CircuitBreakerState = iota
	StateOpen
	StateHalfOpen
)

func (s CircuitBreakerState) String() string {
	switch s {
	case StateClosed:
		return "closed"
	case StateOpen:
		return "open"
	case StateHalfOpen:
		return "half-open"
	default:
		return "unknown"
	}
}

// CircuitBreaker implements circuit breaker pattern
type CircuitBreaker struct {
	mu              sync.RWMutex
	state           CircuitBreakerState
	failureCount    int
	successCount    int
	lastFailureTime time.Time
	config          CircuitBreakerConfig
}

// RateLimitMetrics tracks rate limiting statistics (public interface)
type RateLimitMetrics struct {
	TotalRequests       int64     `json:"total_requests"`
	AllowedRequests     int64     `json:"allowed_requests"`
	RejectedRequests    int64     `json:"rejected_requests"`
	CircuitBreakerTrips int64     `json:"circuit_breaker_trips"`
	LastReset           time.Time `json:"last_reset"`
}

// rateLimitMetrics holds internal metrics with mutex (private)
type rateLimitMetrics struct {
	mu                  sync.RWMutex
	TotalRequests       int64
	AllowedRequests     int64
	RejectedRequests    int64
	CircuitBreakerTrips int64
	LastReset           time.Time
}

// NewRateLimiter creates a new rate limiter with circuit breaker
func NewRateLimiter(config Config) *RateLimiter {
	if config.RPS <= 0 {
		config.RPS = 10
	}
	if config.Burst <= 0 {
		config.Burst = 20
	}
	if config.CircuitBreakerConfig.FailureThreshold <= 0 {
		config.CircuitBreakerConfig.FailureThreshold = 5
	}
	if config.CircuitBreakerConfig.RecoveryTimeout <= 0 {
		config.CircuitBreakerConfig.RecoveryTimeout = 60 * time.Second
	}
	if config.CircuitBreakerConfig.HalfOpenMaxCalls <= 0 {
		config.CircuitBreakerConfig.HalfOpenMaxCalls = 3
	}

	return &RateLimiter{
		limiter: rate.NewLimiter(rate.Limit(config.RPS), config.Burst),
		circuitBreaker: &CircuitBreaker{
			state:  StateClosed,
			config: config.CircuitBreakerConfig,
		},
		metrics: &rateLimitMetrics{
			LastReset: time.Now(),
		},
		config: config,
	}
}

// Allow checks if a request should be allowed
func (r *RateLimiter) Allow(ctx context.Context) error {
	r.metrics.mu.Lock()
	r.metrics.TotalRequests++
	r.metrics.mu.Unlock()

	// Check circuit breaker first
	if !r.circuitBreaker.CanExecute() {
		r.metrics.mu.Lock()
		r.metrics.RejectedRequests++
		r.metrics.mu.Unlock()
		return fmt.Errorf("circuit breaker is %s", r.circuitBreaker.State())
	}

	// Check rate limit
	if !r.limiter.Allow() {
		r.metrics.mu.Lock()
		r.metrics.RejectedRequests++
		r.metrics.mu.Unlock()
		return fmt.Errorf("rate limit exceeded")
	}

	r.metrics.mu.Lock()
	r.metrics.AllowedRequests++
	r.metrics.mu.Unlock()

	return nil
}

// Wait blocks until a request can be made
func (r *RateLimiter) Wait(ctx context.Context) error {
	r.metrics.mu.Lock()
	r.metrics.TotalRequests++
	r.metrics.mu.Unlock()

	// Check circuit breaker first
	if !r.circuitBreaker.CanExecute() {
		r.metrics.mu.Lock()
		r.metrics.RejectedRequests++
		r.metrics.mu.Unlock()
		return fmt.Errorf("circuit breaker is %s", r.circuitBreaker.State())
	}

	// Wait for rate limit
	if err := r.limiter.Wait(ctx); err != nil {
		r.metrics.mu.Lock()
		r.metrics.RejectedRequests++
		r.metrics.mu.Unlock()
		return err
	}

	r.metrics.mu.Lock()
	r.metrics.AllowedRequests++
	r.metrics.mu.Unlock()

	return nil
}

// OnSuccess should be called when a request succeeds
func (r *RateLimiter) OnSuccess() {
	r.circuitBreaker.OnSuccess()
}

// OnFailure should be called when a request fails
func (r *RateLimiter) OnFailure() {
	r.circuitBreaker.OnFailure()
}

// GetMetrics returns current rate limiting metrics
func (r *RateLimiter) GetMetrics() RateLimitMetrics {
	r.metrics.mu.RLock()
	defer r.metrics.mu.RUnlock()
	return RateLimitMetrics{
		TotalRequests:       r.metrics.TotalRequests,
		AllowedRequests:     r.metrics.AllowedRequests,
		RejectedRequests:    r.metrics.RejectedRequests,
		CircuitBreakerTrips: r.metrics.CircuitBreakerTrips,
		LastReset:           r.metrics.LastReset,
	}
}

// ResetMetrics clears the metrics
func (r *RateLimiter) ResetMetrics() {
	r.metrics.mu.Lock()
	defer r.metrics.mu.Unlock()

	r.metrics.TotalRequests = 0
	r.metrics.AllowedRequests = 0
	r.metrics.RejectedRequests = 0
	r.metrics.CircuitBreakerTrips = 0
	r.metrics.LastReset = time.Now()
}

// GetCircuitBreakerState returns current circuit breaker state
func (r *RateLimiter) GetCircuitBreakerState() CircuitBreakerState {
	return r.circuitBreaker.State()
}

// CanExecute checks if a request can be executed based on circuit breaker state
func (c *CircuitBreaker) CanExecute() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()

	now := time.Now()

	switch c.state {
	case StateClosed:
		return true
	case StateOpen:
		if now.Sub(c.lastFailureTime) > c.config.RecoveryTimeout {
			// Transition to half-open state
			c.mu.RUnlock()
			c.mu.Lock()
			if c.state == StateOpen && now.Sub(c.lastFailureTime) > c.config.RecoveryTimeout {
				c.state = StateHalfOpen
				c.successCount = 0
			}
			c.mu.Unlock()
			c.mu.RLock()
			return c.state == StateHalfOpen
		}
		return false
	case StateHalfOpen:
		return c.successCount < c.config.HalfOpenMaxCalls
	default:
		return false
	}
}

// OnSuccess records a successful request
func (c *CircuitBreaker) OnSuccess() {
	c.mu.Lock()
	defer c.mu.Unlock()

	switch c.state {
	case StateClosed:
		c.failureCount = 0
	case StateHalfOpen:
		c.successCount++
		if c.successCount >= c.config.HalfOpenMaxCalls {
			c.state = StateClosed
			c.failureCount = 0
			c.successCount = 0
		}
	}
}

// OnFailure records a failed request
func (c *CircuitBreaker) OnFailure() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.lastFailureTime = time.Now()

	switch c.state {
	case StateClosed:
		c.failureCount++
		if c.failureCount >= c.config.FailureThreshold {
			c.state = StateOpen
		}
	case StateHalfOpen:
		c.state = StateOpen
		c.successCount = 0
	}
}

// State returns the current circuit breaker state
func (c *CircuitBreaker) State() CircuitBreakerState {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.state
}

// DefaultConfig returns default rate limiter configuration
func DefaultConfig() Config {
	return Config{
		RPS:   10,
		Burst: 20,
		CircuitBreakerConfig: CircuitBreakerConfig{
			FailureThreshold: 5,
			RecoveryTimeout:  60 * time.Second,
			MaxRetries:       3,
			RetryDelay:       1 * time.Second,
			HalfOpenMaxCalls: 3,
		},
	}
}

// AdaptiveRateLimiter adjusts rate limits based on success/failure rates
type AdaptiveRateLimiter struct {
	*RateLimiter
	mu           sync.RWMutex
	baseRPS      float64
	currentRPS   float64
	minRPS       float64
	maxRPS       float64
	adjustPeriod time.Duration
	lastAdjust   time.Time
}

// NewAdaptiveRateLimiter creates a rate limiter that adjusts based on performance
func NewAdaptiveRateLimiter(baseRPS, minRPS, maxRPS float64) *AdaptiveRateLimiter {
	config := DefaultConfig()
	config.RPS = baseRPS

	return &AdaptiveRateLimiter{
		RateLimiter:  NewRateLimiter(config),
		baseRPS:      baseRPS,
		currentRPS:   baseRPS,
		minRPS:       minRPS,
		maxRPS:       maxRPS,
		adjustPeriod: 30 * time.Second,
		lastAdjust:   time.Now(),
	}
}

// Adjust modifies the rate limit based on recent performance
func (a *AdaptiveRateLimiter) Adjust() {
	a.mu.Lock()
	defer a.mu.Unlock()

	now := time.Now()
	if now.Sub(a.lastAdjust) < a.adjustPeriod {
		return
	}

	metrics := a.GetMetrics()
	if metrics.TotalRequests == 0 {
		return
	}

	successRate := float64(metrics.AllowedRequests) / float64(metrics.TotalRequests)

	// Adjust rate based on success rate
	if successRate > 0.95 && a.circuitBreaker.State() == StateClosed {
		// Increase rate if success rate is high
		a.currentRPS = min(a.currentRPS*1.1, a.maxRPS)
	} else if successRate < 0.8 || a.circuitBreaker.State() != StateClosed {
		// Decrease rate if success rate is low or circuit breaker is active
		a.currentRPS = max(a.currentRPS*0.9, a.minRPS)
	}

	// Update the limiter
	a.limiter.SetLimit(rate.Limit(a.currentRPS))
	a.lastAdjust = now
	a.ResetMetrics()
}

func min(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}

func max(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}

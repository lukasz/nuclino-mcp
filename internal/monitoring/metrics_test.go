package monitoring

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/lukasz/nuclino-mcp-server/internal/cache"
	"github.com/lukasz/nuclino-mcp-server/internal/ratelimit"
	"github.com/stretchr/testify/assert"
)

func TestMetricsCollector_RecordRequest(t *testing.T) {
	collector := NewMetricsCollector()

	// Record successful requests
	collector.RecordRequest(true, 100*time.Millisecond, "")
	collector.RecordRequest(true, 200*time.Millisecond, "")

	// Record failed request
	collector.RecordRequest(false, 50*time.Millisecond, "timeout")

	metrics := collector.GetMetrics()

	assert.Equal(t, int64(3), metrics.Server.RequestsTotal)
	assert.Equal(t, int64(2), metrics.Server.RequestsSuccessful)
	assert.Equal(t, int64(1), metrics.Server.RequestsFailed)
	assert.Equal(t, int64(1), metrics.Server.ErrorsByType["timeout"])
	assert.Greater(t, metrics.Server.ResponseTimeAvg, time.Duration(0))
}

func TestMetricsCollector_RecordToolCall(t *testing.T) {
	collector := NewMetricsCollector()

	// Record tool calls
	collector.RecordToolCall("test_tool", true, 50*time.Millisecond)
	collector.RecordToolCall("test_tool", true, 100*time.Millisecond)
	collector.RecordToolCall("test_tool", false, 25*time.Millisecond)

	metrics := collector.GetMetrics()

	toolMetric := metrics.Tools["test_tool"]
	assert.NotNil(t, toolMetric)
	assert.Equal(t, int64(3), toolMetric.CallsTotal)
	assert.Equal(t, int64(2), toolMetric.CallsSuccessful)
	assert.Equal(t, int64(1), toolMetric.CallsFailed)
	assert.Equal(t, 25*time.Millisecond, toolMetric.MinLatency)
	assert.Equal(t, 100*time.Millisecond, toolMetric.MaxLatency)
	assert.Greater(t, toolMetric.ErrorRate, float64(0))
}

func TestMetricsCollector_ActiveConnections(t *testing.T) {
	collector := NewMetricsCollector()

	// Test increment
	collector.IncrementActiveConnections()
	collector.IncrementActiveConnections()

	metrics := collector.GetMetrics()
	assert.Equal(t, int64(2), metrics.Server.ActiveConnections)

	// Test decrement
	collector.DecrementActiveConnections()

	metrics = collector.GetMetrics()
	assert.Equal(t, int64(1), metrics.Server.ActiveConnections)

	// Test decrement doesn't go negative
	collector.DecrementActiveConnections()
	collector.DecrementActiveConnections()

	metrics = collector.GetMetrics()
	assert.Equal(t, int64(0), metrics.Server.ActiveConnections)
}

func TestMetricsCollector_WithCache(t *testing.T) {
	collector := NewMetricsCollector()
	cache := cache.NewCache(10, 1*time.Minute)

	collector.SetCache(cache)

	// Generate some cache activity
	cache.Set("key1", "value1")
	cache.Get("key1") // Hit
	cache.Get("key2") // Miss

	metrics := collector.GetMetrics()

	assert.NotNil(t, metrics.Cache)
	assert.Equal(t, int64(1), metrics.Cache.Hits)
	assert.Equal(t, int64(1), metrics.Cache.Misses)
	assert.Equal(t, 50.0, metrics.Cache.HitRate)
	assert.Equal(t, 1, metrics.Cache.Size)
}

func TestMetricsCollector_WithRateLimit(t *testing.T) {
	collector := NewMetricsCollector()
	config := ratelimit.DefaultConfig()
	rateLimiter := ratelimit.NewRateLimiter(config)

	collector.SetRateLimiter(rateLimiter)

	// Generate some rate limiting activity
	ctx := context.Background()
	rateLimiter.Allow(ctx)
	rateLimiter.Allow(ctx)
	rateLimiter.OnSuccess()

	metrics := collector.GetMetrics()

	assert.NotNil(t, metrics.RateLimit)
	assert.Greater(t, metrics.RateLimit.TotalRequests, int64(0))
	assert.Equal(t, "closed", metrics.RateLimit.CircuitBreakerState)
}

func TestMetricsCollector_HealthCheck(t *testing.T) {
	collector := NewMetricsCollector()

	// Healthy scenario
	for i := 0; i < 10; i++ {
		collector.RecordRequest(true, 100*time.Millisecond, "")
	}

	health := collector.HealthCheck()
	assert.True(t, health.Healthy)
	assert.True(t, health.Checks["error_rate"].Healthy)
	assert.True(t, health.Checks["response_time"].Healthy)

	// Unhealthy scenario - high error rate
	for i := 0; i < 5; i++ {
		collector.RecordRequest(false, 100*time.Millisecond, "error")
	}

	health = collector.HealthCheck()
	assert.False(t, health.Healthy)
	assert.False(t, health.Checks["error_rate"].Healthy)
}

func TestMetricsCollector_HealthCheck_HighLatency(t *testing.T) {
	collector := NewMetricsCollector()

	// High latency scenario
	collector.RecordRequest(true, 6*time.Second, "")

	health := collector.HealthCheck()
	assert.False(t, health.Healthy)
	assert.False(t, health.Checks["response_time"].Healthy)
}

func TestMetricsCollector_Reset(t *testing.T) {
	collector := NewMetricsCollector()

	// Add some metrics
	collector.RecordRequest(true, 100*time.Millisecond, "")
	collector.RecordToolCall("test_tool", true, 50*time.Millisecond)
	collector.IncrementActiveConnections()

	// Verify metrics exist
	metrics := collector.GetMetrics()
	assert.Greater(t, metrics.Server.RequestsTotal, int64(0))
	assert.NotEmpty(t, metrics.Tools)

	// Reset
	collector.Reset()

	// Verify metrics are reset
	metrics = collector.GetMetrics()
	assert.Equal(t, int64(0), metrics.Server.RequestsTotal)
	assert.Empty(t, metrics.Tools)
	assert.Equal(t, int64(0), metrics.Server.ActiveConnections)
}

func TestMetricsCollector_ToJSON(t *testing.T) {
	collector := NewMetricsCollector()

	collector.RecordRequest(true, 100*time.Millisecond, "")
	collector.RecordToolCall("test_tool", true, 50*time.Millisecond)

	jsonData, err := collector.ToJSON()

	assert.NoError(t, err)
	assert.NotEmpty(t, jsonData)
	assert.Contains(t, string(jsonData), "server")
	assert.Contains(t, string(jsonData), "tools")
	assert.Contains(t, string(jsonData), "test_tool")
}

func TestMetricsCollector_ResponseTimePercentiles(t *testing.T) {
	collector := NewMetricsCollector()

	// Add enough data points for percentile calculation
	for i := 0; i < 100; i++ {
		collector.RecordRequest(true, time.Duration(i)*time.Millisecond, "")
	}

	metrics := collector.GetMetrics()

	assert.Greater(t, metrics.Server.ResponseTimeP95, time.Duration(0))
	assert.Greater(t, metrics.Server.ResponseTimeP99, time.Duration(0))
	assert.Greater(t, metrics.Server.ResponseTimeP99, metrics.Server.ResponseTimeP95)
}

func TestMetricsCollector_Concurrent(t *testing.T) {
	collector := NewMetricsCollector()

	const numGoroutines = 10
	const numOperations = 100

	var wg sync.WaitGroup

	// Start concurrent goroutines recording metrics
	for g := 0; g < numGoroutines; g++ {
		wg.Add(1)
		go func(goroutineID int) {
			defer wg.Done()

			for i := 0; i < numOperations; i++ {
				collector.RecordRequest(i%2 == 0, time.Duration(i)*time.Millisecond, "")
				collector.RecordToolCall(fmt.Sprintf("tool_%d", goroutineID), true, time.Duration(i)*time.Millisecond)

				if i%10 == 0 {
					collector.IncrementActiveConnections()
				}
				if i%15 == 0 {
					collector.DecrementActiveConnections()
				}
			}
		}(g)
	}

	wg.Wait()

	metrics := collector.GetMetrics()

	// Verify concurrent operations worked
	assert.Equal(t, int64(numGoroutines*numOperations), metrics.Server.RequestsTotal)
	assert.Len(t, metrics.Tools, numGoroutines)
	assert.GreaterOrEqual(t, metrics.Server.ActiveConnections, int64(0))
}

func TestMetricsCollector_MemoryBounds(t *testing.T) {
	collector := NewMetricsCollector()

	// Record many requests to test memory bounds
	for i := 0; i < 2000; i++ {
		collector.RecordRequest(true, time.Duration(i)*time.Millisecond, "")
		collector.RecordToolCall("test_tool", true, time.Duration(i)*time.Millisecond)
	}

	metrics := collector.GetMetrics()

	// Verify metrics were recorded
	assert.Equal(t, int64(2000), metrics.Server.RequestsTotal)
	assert.Equal(t, int64(2000), metrics.Tools["test_tool"].CallsTotal)

	// Memory bounds should prevent unlimited growth
	// (Internal slices should be truncated to reasonable sizes)
	assert.Less(t, len(collector.serverMetrics.responseTimes), 1100) // Should be <= 1000 + some buffer
	assert.Less(t, len(collector.toolMetrics["test_tool"].latencies), 1100)
}

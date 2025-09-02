package monitoring

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/lukasz/nuclino-mcp-server/internal/cache"
	"github.com/lukasz/nuclino-mcp-server/internal/ratelimit"
)

// MetricsCollector collects and aggregates metrics from various components
type MetricsCollector struct {
	mu            sync.RWMutex
	startTime     time.Time
	serverMetrics ServerMetrics
	toolMetrics   map[string]*ToolMetrics
	cache         *cache.Cache
	rateLimiter   *ratelimit.RateLimiter
}

// ServerMetrics tracks overall server performance
type ServerMetrics struct {
	RequestsTotal      int64            `json:"requests_total"`
	RequestsSuccessful int64            `json:"requests_successful"`
	RequestsFailed     int64            `json:"requests_failed"`
	ResponseTimeAvg    time.Duration    `json:"response_time_avg"`
	ResponseTimeP95    time.Duration    `json:"response_time_p95"`
	ResponseTimeP99    time.Duration    `json:"response_time_p99"`
	ActiveConnections  int64            `json:"active_connections"`
	ErrorsByType       map[string]int64 `json:"errors_by_type"`
	Uptime             time.Duration    `json:"uptime"`
	responseTimes      []time.Duration  // Internal for percentile calculations
}

// ToolMetrics tracks metrics for individual MCP tools
type ToolMetrics struct {
	Name            string          `json:"name"`
	CallsTotal      int64           `json:"calls_total"`
	CallsSuccessful int64           `json:"calls_successful"`
	CallsFailed     int64           `json:"calls_failed"`
	AverageLatency  time.Duration   `json:"average_latency"`
	MaxLatency      time.Duration   `json:"max_latency"`
	MinLatency      time.Duration   `json:"min_latency"`
	LastCalled      time.Time       `json:"last_called"`
	ErrorRate       float64         `json:"error_rate"`
	latencies       []time.Duration // Internal for calculations
}

// SystemMetrics provides system-level metrics
type SystemMetrics struct {
	CPUUsage       float64 `json:"cpu_usage"`
	MemoryUsed     int64   `json:"memory_used_bytes"`
	MemoryTotal    int64   `json:"memory_total_bytes"`
	GoroutineCount int     `json:"goroutine_count"`
	GCPauses       int64   `json:"gc_pauses_total"`
}

// CombinedMetrics aggregates all metrics
type CombinedMetrics struct {
	Server    ServerMetrics           `json:"server"`
	Tools     map[string]*ToolMetrics `json:"tools"`
	Cache     *CacheMetrics           `json:"cache,omitempty"`
	RateLimit *RateLimitMetrics       `json:"rate_limit,omitempty"`
	System    SystemMetrics           `json:"system"`
	Timestamp time.Time               `json:"timestamp"`
}

// CacheMetrics wraps cache statistics
type CacheMetrics struct {
	Hits        int64   `json:"hits"`
	Misses      int64   `json:"misses"`
	Evictions   int64   `json:"evictions"`
	Expirations int64   `json:"expirations"`
	HitRate     float64 `json:"hit_rate"`
	Size        int     `json:"current_size"`
}

// RateLimitMetrics wraps rate limiter statistics
type RateLimitMetrics struct {
	TotalRequests       int64   `json:"total_requests"`
	AllowedRequests     int64   `json:"allowed_requests"`
	RejectedRequests    int64   `json:"rejected_requests"`
	CircuitBreakerTrips int64   `json:"circuit_breaker_trips"`
	AllowedRate         float64 `json:"allowed_rate"`
	CircuitBreakerState string  `json:"circuit_breaker_state"`
}

// NewMetricsCollector creates a new metrics collector
func NewMetricsCollector() *MetricsCollector {
	return &MetricsCollector{
		startTime: time.Now(),
		serverMetrics: ServerMetrics{
			ErrorsByType: make(map[string]int64),
		},
		toolMetrics: make(map[string]*ToolMetrics),
	}
}

// SetCache sets the cache instance for metrics collection
func (m *MetricsCollector) SetCache(c *cache.Cache) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.cache = c
}

// SetRateLimiter sets the rate limiter instance for metrics collection
func (m *MetricsCollector) SetRateLimiter(rl *ratelimit.RateLimiter) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.rateLimiter = rl
}

// RecordRequest records a server request
func (m *MetricsCollector) RecordRequest(success bool, responseTime time.Duration, errorType string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.serverMetrics.RequestsTotal++
	if success {
		m.serverMetrics.RequestsSuccessful++
	} else {
		m.serverMetrics.RequestsFailed++
		if errorType != "" {
			m.serverMetrics.ErrorsByType[errorType]++
		}
	}

	// Update response times
	m.serverMetrics.responseTimes = append(m.serverMetrics.responseTimes, responseTime)

	// Keep only recent response times to avoid memory growth
	if len(m.serverMetrics.responseTimes) > 1000 {
		m.serverMetrics.responseTimes = m.serverMetrics.responseTimes[len(m.serverMetrics.responseTimes)-1000:]
	}

	// Update average response time
	m.updateResponseTimeStats()
}

// RecordToolCall records a tool execution
func (m *MetricsCollector) RecordToolCall(toolName string, success bool, latency time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.toolMetrics[toolName] == nil {
		m.toolMetrics[toolName] = &ToolMetrics{
			Name:       toolName,
			MinLatency: latency,
			MaxLatency: latency,
		}
	}

	tool := m.toolMetrics[toolName]
	tool.CallsTotal++
	tool.LastCalled = time.Now()

	if success {
		tool.CallsSuccessful++
	} else {
		tool.CallsFailed++
	}

	// Update latency stats
	tool.latencies = append(tool.latencies, latency)
	if len(tool.latencies) > 1000 {
		tool.latencies = tool.latencies[len(tool.latencies)-1000:]
	}

	if latency < tool.MinLatency {
		tool.MinLatency = latency
	}
	if latency > tool.MaxLatency {
		tool.MaxLatency = latency
	}

	// Update average latency
	m.updateToolStats(tool)
}

// IncrementActiveConnections increments active connection count
func (m *MetricsCollector) IncrementActiveConnections() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.serverMetrics.ActiveConnections++
}

// DecrementActiveConnections decrements active connection count
func (m *MetricsCollector) DecrementActiveConnections() {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.serverMetrics.ActiveConnections > 0 {
		m.serverMetrics.ActiveConnections--
	}
}

// GetMetrics returns current metrics snapshot
func (m *MetricsCollector) GetMetrics() CombinedMetrics {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Create a deep copy to avoid concurrent access issues
	serverMetrics := m.serverMetrics
	serverMetrics.Uptime = time.Since(m.startTime)
	serverMetrics.ErrorsByType = make(map[string]int64)
	for k, v := range m.serverMetrics.ErrorsByType {
		serverMetrics.ErrorsByType[k] = v
	}

	toolMetrics := make(map[string]*ToolMetrics)
	for k, v := range m.toolMetrics {
		toolCopy := *v
		toolMetrics[k] = &toolCopy
	}

	combined := CombinedMetrics{
		Server:    serverMetrics,
		Tools:     toolMetrics,
		System:    m.getSystemMetrics(),
		Timestamp: time.Now(),
	}

	// Add cache metrics if available
	if m.cache != nil {
		stats := m.cache.Stats()
		combined.Cache = &CacheMetrics{
			Hits:        stats.Hits,
			Misses:      stats.Misses,
			Evictions:   stats.Evictions,
			Expirations: stats.Expirations,
			HitRate:     stats.HitRate(),
			Size:        m.cache.Size(),
		}
	}

	// Add rate limit metrics if available
	if m.rateLimiter != nil {
		rateLimitStats := m.rateLimiter.GetMetrics()
		allowedRate := float64(0)
		if rateLimitStats.TotalRequests > 0 {
			allowedRate = float64(rateLimitStats.AllowedRequests) / float64(rateLimitStats.TotalRequests) * 100
		}

		combined.RateLimit = &RateLimitMetrics{
			TotalRequests:       rateLimitStats.TotalRequests,
			AllowedRequests:     rateLimitStats.AllowedRequests,
			RejectedRequests:    rateLimitStats.RejectedRequests,
			CircuitBreakerTrips: rateLimitStats.CircuitBreakerTrips,
			AllowedRate:         allowedRate,
			CircuitBreakerState: m.rateLimiter.GetCircuitBreakerState().String(),
		}
	}

	return combined
}

// updateResponseTimeStats calculates percentiles and averages
func (m *MetricsCollector) updateResponseTimeStats() {
	if len(m.serverMetrics.responseTimes) == 0 {
		return
	}

	// Calculate average
	var total time.Duration
	for _, rt := range m.serverMetrics.responseTimes {
		total += rt
	}
	m.serverMetrics.ResponseTimeAvg = total / time.Duration(len(m.serverMetrics.responseTimes))

	// For simplicity, we'll approximate percentiles
	// In production, you'd want a proper percentile calculation
	if len(m.serverMetrics.responseTimes) >= 20 {
		sorted := make([]time.Duration, len(m.serverMetrics.responseTimes))
		copy(sorted, m.serverMetrics.responseTimes)

		// Simple bubble sort for demonstration
		for i := 0; i < len(sorted); i++ {
			for j := i + 1; j < len(sorted); j++ {
				if sorted[i] > sorted[j] {
					sorted[i], sorted[j] = sorted[j], sorted[i]
				}
			}
		}

		p95Index := int(float64(len(sorted)) * 0.95)
		p99Index := int(float64(len(sorted)) * 0.99)

		if p95Index < len(sorted) {
			m.serverMetrics.ResponseTimeP95 = sorted[p95Index]
		}
		if p99Index < len(sorted) {
			m.serverMetrics.ResponseTimeP99 = sorted[p99Index]
		}
	}
}

// updateToolStats updates tool-specific statistics
func (m *MetricsCollector) updateToolStats(tool *ToolMetrics) {
	if len(tool.latencies) == 0 {
		return
	}

	// Calculate average latency
	var total time.Duration
	for _, latency := range tool.latencies {
		total += latency
	}
	tool.AverageLatency = total / time.Duration(len(tool.latencies))

	// Calculate error rate
	if tool.CallsTotal > 0 {
		tool.ErrorRate = float64(tool.CallsFailed) / float64(tool.CallsTotal) * 100
	}
}

// getSystemMetrics returns current system metrics
func (m *MetricsCollector) getSystemMetrics() SystemMetrics {
	// For demonstration, return basic values
	// In production, you'd integrate with system monitoring libraries
	return SystemMetrics{
		CPUUsage:       0, // Would use runtime/pprof or external library
		MemoryUsed:     0, // Would use runtime.MemStats
		MemoryTotal:    0, // Would get from system
		GoroutineCount: 0, // Would use runtime.NumGoroutine()
		GCPauses:       0, // Would use runtime.MemStats
	}
}

// Reset resets all metrics
func (m *MetricsCollector) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.startTime = time.Now()
	m.serverMetrics = ServerMetrics{
		ErrorsByType: make(map[string]int64),
	}
	m.toolMetrics = make(map[string]*ToolMetrics)
}

// ToJSON returns metrics as JSON
func (m *MetricsCollector) ToJSON() ([]byte, error) {
	metrics := m.GetMetrics()
	return json.MarshalIndent(metrics, "", "  ")
}

// HealthCheck performs a health check based on metrics
func (m *MetricsCollector) HealthCheck() HealthStatus {
	metrics := m.GetMetrics()

	status := HealthStatus{
		Healthy:   true,
		Timestamp: time.Now(),
		Checks:    make(map[string]CheckResult),
	}

	// Check error rate
	if metrics.Server.RequestsTotal > 0 {
		errorRate := float64(metrics.Server.RequestsFailed) / float64(metrics.Server.RequestsTotal) * 100
		if errorRate > 10 { // More than 10% error rate
			status.Checks["error_rate"] = CheckResult{
				Healthy: false,
				Message: fmt.Sprintf("High error rate: %.2f%%", errorRate),
			}
			status.Healthy = false
		} else {
			status.Checks["error_rate"] = CheckResult{
				Healthy: true,
				Message: fmt.Sprintf("Error rate: %.2f%%", errorRate),
			}
		}
	}

	// Check response time
	if metrics.Server.ResponseTimeAvg > 5*time.Second {
		status.Checks["response_time"] = CheckResult{
			Healthy: false,
			Message: fmt.Sprintf("High response time: %v", metrics.Server.ResponseTimeAvg),
		}
		status.Healthy = false
	} else {
		status.Checks["response_time"] = CheckResult{
			Healthy: true,
			Message: fmt.Sprintf("Response time: %v", metrics.Server.ResponseTimeAvg),
		}
	}

	// Check circuit breaker
	if metrics.RateLimit != nil && metrics.RateLimit.CircuitBreakerState == "open" {
		status.Checks["circuit_breaker"] = CheckResult{
			Healthy: false,
			Message: "Circuit breaker is open",
		}
		status.Healthy = false
	} else if metrics.RateLimit != nil {
		status.Checks["circuit_breaker"] = CheckResult{
			Healthy: true,
			Message: fmt.Sprintf("Circuit breaker: %s", metrics.RateLimit.CircuitBreakerState),
		}
	}

	// Check cache hit rate
	if metrics.Cache != nil && metrics.Cache.HitRate < 50 {
		status.Checks["cache_performance"] = CheckResult{
			Healthy: false,
			Message: fmt.Sprintf("Low cache hit rate: %.2f%%", metrics.Cache.HitRate),
		}
	} else if metrics.Cache != nil {
		status.Checks["cache_performance"] = CheckResult{
			Healthy: true,
			Message: fmt.Sprintf("Cache hit rate: %.2f%%", metrics.Cache.HitRate),
		}
	}

	return status
}

// HealthStatus represents the overall health of the system
type HealthStatus struct {
	Healthy   bool                   `json:"healthy"`
	Timestamp time.Time              `json:"timestamp"`
	Checks    map[string]CheckResult `json:"checks"`
}

// CheckResult represents the result of an individual health check
type CheckResult struct {
	Healthy bool   `json:"healthy"`
	Message string `json:"message"`
}

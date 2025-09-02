# Enhanced Features Documentation

This document describes the advanced features implemented for the Nuclino MCP Server, including rate limiting, caching, comprehensive error handling, monitoring, and performance optimizations.

## 🚀 Overview

The enhanced Nuclino MCP Server includes:

- **Advanced Rate Limiting** with circuit breaker pattern
- **Intelligent Caching** with TTL and LRU eviction
- **Comprehensive Error Handling** with automatic retries
- **Performance Monitoring** and metrics collection
- **Enhanced Client** with all features integrated
- **Stress Testing** and performance benchmarks

## 📊 Performance Metrics

Based on benchmarks on Apple M1 Pro:

- **Cache Set Operations**: ~9,600 ns/op
- **Cache Get Operations**: ~156 ns/op  
- **Mixed Cache Operations**: ~3,400 ns/op
- **Rate Limiter**: ~425 ns/op
- **Cache with Expiration**: ~5,300 ns/op

## 🔒 Rate Limiting & Circuit Breaker

### Features
- **Configurable rate limits** with burst capacity
- **Circuit breaker pattern** for fault tolerance
- **Adaptive rate limiting** based on success rates
- **Comprehensive metrics** tracking

### Configuration
```go
config := ratelimit.Config{
    RPS:   10,    // Requests per second
    Burst: 20,    // Burst capacity
    CircuitBreakerConfig: ratelimit.CircuitBreakerConfig{
        FailureThreshold: 5,              // Failures before opening
        RecoveryTimeout:  60 * time.Second, // Time before half-open
        HalfOpenMaxCalls: 3,              // Calls allowed in half-open
    },
}
```

### Usage
```go
rateLimiter := ratelimit.NewRateLimiter(config)

// Check if request is allowed
if err := rateLimiter.Allow(ctx); err != nil {
    // Handle rate limit or circuit breaker rejection
}

// Record success/failure for circuit breaker
rateLimiter.OnSuccess()
rateLimiter.OnFailure()
```

### Circuit Breaker States
- **Closed**: Normal operation, requests allowed
- **Open**: Circuit tripped, requests rejected  
- **Half-Open**: Testing recovery, limited requests allowed

## 💾 Intelligent Caching

### Features
- **TTL-based expiration** with custom timeouts
- **LRU eviction** when capacity is exceeded
- **Concurrent access** with proper locking
- **Hit/miss metrics** and performance tracking
- **Memory-bounded** to prevent unlimited growth

### Configuration
```go
cacheConfig := cache.CacheConfig{
    MaxSize:       1000,
    DefaultTTL:    5 * time.Minute,
    ItemTTL:       10 * time.Minute,
    WorkspaceTTL:  30 * time.Minute,
    CollectionTTL: 15 * time.Minute,
    SearchTTL:     2 * time.Minute,
}
```

### Usage
```go
cache := cache.NewCache(1000, 5*time.Minute)

// Set with default TTL
cache.Set("key", "value")

// Set with custom TTL
cache.SetWithTTL("key", "value", 1*time.Minute)

// Get value
if value, found := cache.Get("key"); found {
    // Use cached value
}

// Get statistics
stats := cache.Stats()
fmt.Printf("Hit rate: %.2f%%", stats.HitRate())
```

### Cache Statistics
- **Hits/Misses**: Request success metrics
- **Evictions**: LRU removals due to capacity
- **Expirations**: TTL-based removals
- **Hit Rate**: Percentage of successful lookups

## ⚠️ Enhanced Error Handling

### Error Types
- **Validation**: Input parameter errors
- **Authentication**: API key issues
- **Authorization**: Permission denied
- **Rate Limit**: Quota exceeded
- **Network**: Connection problems
- **Timeout**: Request timeouts
- **Circuit Breaker**: Service protection
- **API**: Server-side errors
- **Internal**: System errors

### Error Structure
```go
type Error struct {
    Type         ErrorType              // Category of error
    Code         string                 // Specific error code
    Message      string                 // Human-readable message
    Details      string                 // Additional context
    HTTPStatus   int                    // HTTP status code
    Severity     Severity               // Low/Medium/High/Critical
    Retryable    bool                   // Whether retry is recommended
    Timestamp    time.Time              // When error occurred
    Context      map[string]interface{} // Additional metadata
}
```

### Retry Configuration
```go
retryConfig := errors.RetryConfig{
    MaxRetries:    3,
    InitialDelay:  1 * time.Second,
    MaxDelay:      30 * time.Second,
    BackoffFactor: 2.0,
    RetryableErrors: []ErrorType{
        errors.ErrorTypeNetwork,
        errors.ErrorTypeTimeout,
        errors.ErrorTypeRateLimit,
    },
}
```

### Usage Examples
```go
// Create specific error types
validationErr := errors.NewValidationError("email", "invalid format")
authErr := errors.NewAuthenticationError("invalid API key")
rateLimitErr := errors.NewRateLimitError(time.Now().Add(time.Minute))

// Check error properties
if errors.IsRetryable(err) {
    // Implement retry logic
}

httpStatus := errors.GetHTTPStatus(err)
errorType := errors.GetErrorType(err)
```

## 📈 Monitoring & Metrics

### Server Metrics
- **Request Statistics**: Total, successful, failed requests
- **Response Times**: Average, P95, P99 percentiles
- **Active Connections**: Current connection count
- **Error Breakdown**: Errors by type
- **Uptime**: Server uptime duration

### Tool Metrics (Per MCP Tool)
- **Call Statistics**: Total, successful, failed calls
- **Latency Metrics**: Average, min, max latency
- **Error Rate**: Percentage of failed calls
- **Last Called**: Timestamp of most recent call

### System Metrics
- **Cache Performance**: Hit rate, size, evictions
- **Rate Limiting**: Allowed/rejected requests, circuit breaker state
- **Memory Usage**: Current and total memory (extensible)
- **Goroutines**: Active goroutine count (extensible)

### Usage
```go
// Create metrics collector
collector := monitoring.NewMetricsCollector()

// Set cache and rate limiter for monitoring
collector.SetCache(cache)
collector.SetRateLimiter(rateLimiter)

// Record metrics
collector.RecordRequest(true, 100*time.Millisecond, "")
collector.RecordToolCall("nuclino_get_item", true, 50*time.Millisecond)
collector.IncrementActiveConnections()

// Get current metrics
metrics := collector.GetMetrics()

// Export as JSON
jsonData, err := collector.ToJSON()

// Health check
health := collector.HealthCheck()
if !health.Healthy {
    // Handle unhealthy state
}
```

### Health Checks
- **Error Rate**: Alerts if > 10% error rate
- **Response Time**: Alerts if average > 5 seconds
- **Circuit Breaker**: Alerts if circuit is open
- **Cache Performance**: Alerts if hit rate < 50%

## 🎯 Enhanced Client Integration

The `EnhancedClient` integrates all features:

```go
// Configuration
config := nuclino.EnhancedClientConfig{
    APIKey:          "your-api-key",
    BaseURL:         "https://api.nuclino.com",
    Timeout:         30 * time.Second,
    RetryConfig:     errors.DefaultRetryConfig(),
    RateLimitConfig: ratelimit.DefaultConfig(),
    CacheConfig:     cache.DefaultCacheConfig(),
    EnableCache:     true,
    EnableMetrics:   true,
}

// Create client with logger
logger := &MyLogger{} // Implement errors.Logger interface
client := nuclino.NewEnhancedClient(config, logger)

// All API calls automatically include:
// - Rate limiting with circuit breaker
// - Intelligent caching (for GET requests)
// - Comprehensive error handling with retries
// - Performance metrics collection

user, err := client.GetCurrentUser(ctx)
if err != nil {
    // Error is already categorized and logged
}

// Get performance metrics
metrics := client.GetMetrics()
cacheStats := client.GetCacheStats()
rateLimitMetrics := client.GetRateLimiterMetrics()
```

## 🧪 Testing & Performance

### Test Coverage
- **Unit Tests**: All components with mock dependencies
- **Integration Tests**: Complex workflows and component interaction
- **Stress Tests**: High-concurrency scenarios
- **Performance Tests**: Memory bounds and resource usage
- **Benchmarks**: Throughput and latency measurements

### Running Tests
```bash
# All tests
go test -v ./internal/cache ./internal/ratelimit ./internal/monitoring

# Performance tests
go test -v ./internal/performance

# Benchmarks
go test -bench=. ./internal/performance

# Stress tests with race detection
go test -race -v ./internal/performance
```

### Performance Results
Based on stress testing:

- **Cache**: Handles 100k+ concurrent operations
- **Rate Limiter**: Processes 2.8M+ operations/second
- **Memory**: Bounded growth with automatic cleanup
- **Concurrency**: Thread-safe across all components

## 🛠️ Configuration Best Practices

### Rate Limiting
- Set **RPS** based on API quotas (default: 10/sec)
- Configure **Burst** for traffic spikes (default: 20)
- Adjust **Circuit Breaker** thresholds for fault tolerance

### Caching
- Use appropriate **TTL** values per data type:
  - **Workspaces**: 30 minutes (rarely change)
  - **Collections**: 15 minutes (moderate change)
  - **Items**: 10 minutes (frequent change)
  - **Search**: 2 minutes (dynamic results)
- Set **MaxSize** based on available memory

### Error Handling
- Enable **retries** for transient errors
- Use **exponential backoff** to avoid thundering herd
- Set **reasonable timeouts** to prevent hanging

### Monitoring
- Track **error rates** and **response times**
- Monitor **cache hit rates** for effectiveness
- Watch **circuit breaker** state for service health
- Set up **alerts** for health check failures

## 🔧 Troubleshooting

### Common Issues

1. **High Error Rate**
   - Check API key validity
   - Verify network connectivity
   - Review rate limiting settings

2. **Poor Cache Performance**
   - Increase cache size if memory allows
   - Adjust TTL values for data volatility
   - Monitor hit rates and eviction patterns

3. **Circuit Breaker Opens Frequently**
   - Lower failure threshold if too sensitive
   - Increase recovery timeout for stability
   - Check underlying service health

4. **High Memory Usage**
   - Reduce cache size
   - Lower TTL values for faster expiration
   - Check for memory leaks in custom code

### Debugging Tools

```go
// Enable debug logging
config.EnableDebugLogging = true

// Get detailed metrics
metrics := collector.GetMetrics()
fmt.Printf("Metrics: %+v", metrics)

// Check cache statistics
stats := cache.Stats()
fmt.Printf("Cache hit rate: %.2f%%", stats.HitRate())

// Monitor circuit breaker state
state := rateLimiter.GetCircuitBreakerState()
fmt.Printf("Circuit breaker: %s", state)
```

## 📚 API Reference

### Cache Package
- `NewCache(maxSize, defaultTTL)` - Create cache instance
- `Set(key, value)` - Store value with default TTL
- `SetWithTTL(key, value, ttl)` - Store value with custom TTL
- `Get(key)` - Retrieve value
- `Delete(key)` - Remove value
- `Clear()` - Remove all values
- `Stats()` - Get performance statistics

### Rate Limit Package
- `NewRateLimiter(config)` - Create rate limiter
- `Allow(ctx)` - Check if request allowed
- `Wait(ctx)` - Wait for request slot
- `OnSuccess()` - Record successful request
- `OnFailure()` - Record failed request
- `GetMetrics()` - Get rate limiting statistics

### Monitoring Package
- `NewMetricsCollector()` - Create metrics collector
- `RecordRequest(success, responseTime, errorType)` - Record request
- `RecordToolCall(toolName, success, latency)` - Record tool call
- `GetMetrics()` - Get all metrics
- `HealthCheck()` - Perform health check
- `ToJSON()` - Export metrics as JSON

### Errors Package
- `NewError(type, code, message)` - Create error
- `NewValidationError(field, message)` - Create validation error
- `NewRateLimitError(resetTime)` - Create rate limit error
- `IsRetryable(err)` - Check if error is retryable
- `GetHTTPStatus(err)` - Extract HTTP status code

This enhanced implementation provides enterprise-grade reliability, observability, and performance for the Nuclino MCP Server.
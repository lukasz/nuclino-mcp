# Development Guide

Complete guide for developing and extending the Nuclino MCP Server.

## Development Environment Setup

### Prerequisites
- Go 1.21 or later
- [Mage](https://magefile.org/) - `go install github.com/magefile/mage@latest`
- Git for version control

### Initial Setup

1. **Clone and initialize:**
   ```bash
   git clone <repository>
   cd nuclino-mcp-server
   mage install  # Install dependencies
   ```

2. **Environment configuration:**
   ```bash
   cp .env.example .env
   # Edit .env and add your NUCLINO_API_KEY
   ```

3. **Build and test:**
   ```bash
   mage build    # Build binary
   mage test     # Run all tests
   mage dev      # Run in debug mode
   ```

## Build System (Mage)

### Available Commands

```bash
# Building
mage build        # Build binary for current platform
mage buildall     # Cross-platform builds
mage install      # Install dependencies

# Development
mage run          # Run server
mage dev          # Run with debug logging
mage fmt          # Format code
mage clean        # Clean build artifacts

# Testing
mage test         # Run all tests
mage testcoverage # Tests with coverage report
mage ci           # Full CI pipeline (format, lint, test, build)

# Code Quality
mage lint         # Run golangci-lint
mage vet          # Run go vet
```

### Build Configuration

The build system supports:
- **Cross-compilation:** Linux, macOS, Windows (AMD64, ARM64)
- **Version embedding:** Git commit and build timestamp
- **Optimization:** Stripped binaries for production

## Project Architecture

### Directory Structure
```
cmd/server/              # Application entry point
├── main.go             # Server bootstrap

internal/server/         # MCP server implementation
├── server.go           # Core MCP server
└── handler.go          # Tool request handling

internal/nuclino/        # Nuclino API integration
├── client.go           # Basic HTTP client
├── enhanced_client.go  # Enterprise client with features
└── types.go            # API response types

internal/tools/          # MCP tool implementations
├── registry.go         # Tool registration
├── items.go            # Item management tools
├── workspace_extended.go # Extended workspace tools
├── collections_extended.go # Extended collection tools
└── *_test.go           # Comprehensive tests

internal/cache/          # Intelligent caching system
├── cache.go            # LRU cache with TTL
└── cache_test.go       # Cache tests

internal/ratelimit/      # Advanced rate limiting
├── ratelimit.go        # Circuit breaker pattern
└── ratelimit_test.go   # Rate limiting tests

internal/errors/         # Comprehensive error handling
├── errors.go           # Error types and handling
└── retry.go            # Retry logic

internal/monitoring/     # Performance monitoring
├── metrics.go          # Metrics collection
└── health.go           # Health checks

internal/performance/    # Performance testing
└── performance_test.go # Stress tests and benchmarks
```

## Adding New Tools

### 1. Define the Tool

Create a new tool in `internal/tools/`:

```go
// NewExampleTool creates a new example tool
func NewExampleTool(client nuclino.Client) *ExampleTool {
    return &ExampleTool{client: client}
}

type ExampleTool struct {
    client nuclino.Client
}

func (t *ExampleTool) Name() string {
    return "nuclino_example"
}

func (t *ExampleTool) Description() string {
    return "Example tool description"
}

func (t *ExampleTool) InputSchema() map[string]interface{} {
    return map[string]interface{}{
        "type": "object",
        "properties": map[string]interface{}{
            "example_param": map[string]interface{}{
                "type": "string",
                "description": "Example parameter",
            },
        },
        "required": []string{"example_param"},
    }
}

func (t *ExampleTool) Execute(ctx context.Context, args map[string]interface{}) (interface{}, error) {
    // Extract arguments
    exampleParam, ok := args["example_param"].(string)
    if !ok {
        return nil, errors.NewValidationError("example_param", "must be a string")
    }

    // Implement tool logic
    result, err := t.client.ExampleOperation(ctx, exampleParam)
    if err != nil {
        return nil, err
    }

    return result, nil
}
```

### 2. Add Tests

Create comprehensive tests in `*_test.go`:

```go
func TestExampleTool_Execute(t *testing.T) {
    mockClient := new(MockClient)
    tool := NewExampleTool(mockClient)

    // Test successful execution
    mockClient.On("ExampleOperation", mock.Anything, "test").Return("result", nil)

    result, err := tool.Execute(context.Background(), map[string]interface{}{
        "example_param": "test",
    })

    assert.NoError(t, err)
    assert.Equal(t, "result", result)
    mockClient.AssertExpectations(t)
}

func TestExampleTool_Execute_ValidationError(t *testing.T) {
    tool := NewExampleTool(new(MockClient))

    _, err := tool.Execute(context.Background(), map[string]interface{}{})

    assert.Error(t, err)
    assert.Contains(t, err.Error(), "example_param")
}
```

### 3. Register the Tool

Add to `internal/tools/registry.go`:

```go
func RegisterTools(server *server.MCPServer, client nuclino.Client) {
    // ... existing tools ...
    
    // Register new tool
    server.RegisterTool(NewExampleTool(client))
}
```

## Testing Strategy

### Test Types

1. **Unit Tests:** Individual tool testing with mocks
2. **Integration Tests:** Multi-tool workflows
3. **Performance Tests:** Stress testing and benchmarks
4. **Error Handling Tests:** Edge cases and failures

### Running Tests

```bash
# All tests
mage test

# Specific test suites
go test -v ./internal/tools -run TestExample
go test -v ./internal/tools -run Integration
go test -v ./internal/cache -run TestCache

# With coverage
mage testcoverage

# Performance tests
go test -v ./internal/performance -run Benchmark
```

### Test Data and Mocks

The project uses:
- **testify/mock:** For mocking API clients
- **testify/assert:** For assertions
- **Context handling:** Proper context propagation
- **Error scenarios:** Comprehensive error testing

## Adding Enterprise Features

### Caching Integration

Add caching to your tool:

```go
func (t *ExampleTool) Execute(ctx context.Context, args map[string]interface{}) (interface{}, error) {
    cacheKey := fmt.Sprintf("example:%s", args["example_param"])
    
    // Try cache first
    if cached, found := t.cache.Get(cacheKey); found {
        return cached, nil
    }

    // Execute operation
    result, err := t.client.ExampleOperation(ctx, args["example_param"].(string))
    if err != nil {
        return nil, err
    }

    // Cache result
    t.cache.Set(cacheKey, result, 5*time.Minute)
    
    return result, nil
}
```

### Error Handling

Use the comprehensive error system:

```go
import "github.com/lukasz/nuclino-mcp-server/internal/errors"

// Validation error
if exampleParam == "" {
    return nil, errors.NewValidationError("example_param", "cannot be empty")
}

// API error with context
result, err := t.client.ExampleOperation(ctx, exampleParam)
if err != nil {
    return nil, errors.NewAPIError(500, "EXAMPLE_ERROR", "Example operation failed").
        WithCause(err).
        WithContext("param", exampleParam)
}
```

### Monitoring Integration

Add metrics to your tool:

```go
func (t *ExampleTool) Execute(ctx context.Context, args map[string]interface{}) (interface{}, error) {
    start := time.Now()
    
    result, err := t.performOperation(ctx, args)
    
    // Record metrics
    success := err == nil
    duration := time.Since(start)
    t.metrics.RecordToolCall("example_tool", success, duration)
    
    return result, err
}
```

## API Client Extension

### Adding New Endpoints

Extend the Nuclino client in `internal/nuclino/client.go`:

```go
func (c *Client) ExampleOperation(ctx context.Context, param string) (*ExampleResponse, error) {
    var response ExampleResponse
    
    resp, err := c.httpClient.R().
        SetContext(ctx).
        SetResult(&response).
        SetPathParam("param", param).
        Get("/example/{param}")
    
    if err != nil {
        return nil, errors.NewNetworkError("example_operation", err)
    }
    
    if resp.IsError() {
        return nil, errors.NewAPIError(resp.StatusCode(), "EXAMPLE_ERROR", "Operation failed")
    }
    
    return &response, nil
}
```

### Adding Response Types

Define types in `internal/nuclino/types.go`:

```go
type ExampleResponse struct {
    ID          string    `json:"id"`
    Name        string    `json:"name"`
    Description string    `json:"description"`
    CreatedAt   time.Time `json:"created_at"`
    UpdatedAt   time.Time `json:"updated_at"`
}
```

## Configuration Management

### Environment Variables

Add new configuration options in the appropriate config structure:

```go
type Config struct {
    // ... existing fields ...
    ExampleSetting string        `env:"EXAMPLE_SETTING" envDefault:"default_value"`
    ExampleTimeout time.Duration `env:"EXAMPLE_TIMEOUT" envDefault:"30s"`
}
```

### Validation

Add validation in the config loading:

```go
func LoadConfig() (*Config, error) {
    config := &Config{}
    
    if err := env.Parse(config); err != nil {
        return nil, err
    }
    
    // Validate example setting
    if config.ExampleSetting == "" {
        return nil, errors.New("EXAMPLE_SETTING is required")
    }
    
    return config, nil
}
```

## Performance Optimization

### Benchmarking

Add benchmarks for performance-critical code:

```go
func BenchmarkExampleOperation(b *testing.B) {
    client := setupTestClient()
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        _, err := client.ExampleOperation(context.Background(), "test")
        if err != nil {
            b.Fatal(err)
        }
    }
}
```

### Profiling

Enable profiling for performance analysis:

```bash
# Build with profiling
go build -tags profile

# Run with profiling
./nuclino-mcp-server -cpuprofile=cpu.prof -memprofile=mem.prof

# Analyze profiles
go tool pprof cpu.prof
go tool pprof mem.prof
```

## Debugging and Troubleshooting

### Debug Mode

Enable debug logging:

```bash
mage dev  # Run with debug logging
```

Or set environment:

```bash
LOG_LEVEL=debug DEBUG=true ./nuclino-mcp-server
```

### Common Debug Techniques

1. **Request Tracing:** Add request IDs for tracking
2. **Structured Logging:** Use consistent log format
3. **Error Context:** Include relevant context in errors
4. **Metrics:** Monitor performance and error rates

### Testing API Connectivity

Use the test utility:

```bash
# Test with real API
go run test_mcp_server.go

# Test with mock data  
NUCLINO_API_KEY="" go run test_mcp_server.go
```

## Release Process

### Version Management

1. **Update version** in relevant files
2. **Create git tag:** `git tag v1.2.3`
3. **Build releases:** `mage buildall`
4. **Test builds** on target platforms

### CI/CD Pipeline

The project includes automated checks:

```bash
mage ci  # Runs: format, lint, test, build
```

This ensures:
- Code formatting consistency
- Linting compliance
- All tests passing
- Successful builds

## Contributing Guidelines

### Code Standards

1. **Go Formatting:** Use `gofmt` and `goimports`
2. **Linting:** Pass `golangci-lint` checks
3. **Testing:** Maintain high test coverage
4. **Documentation:** Update relevant docs

### Pull Request Process

1. **Create feature branch:** `git checkout -b feature-name`
2. **Implement changes** with tests
3. **Run CI pipeline:** `mage ci`
4. **Update documentation** if needed
5. **Submit pull request** with clear description

### Testing Requirements

- Unit tests for all new functionality
- Integration tests for complex workflows
- Error handling tests for edge cases
- Performance tests for critical paths

This development guide provides a comprehensive foundation for extending and maintaining the Nuclino MCP Server.
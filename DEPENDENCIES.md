# Dependencies Summary

This document provides an overview of all external libraries and tools used in the Nuclino MCP Server project.

## Core Dependencies

### HTTP Client
- **[github.com/go-resty/resty/v2](https://github.com/go-resty/resty)** `v2.11.0`
  - **Purpose**: HTTP client library for API calls to Nuclino
  - **Features**: Automatic retry, request/response middleware, JSON handling, error handling
  - **Why chosen**: Battle-tested, feature-rich, excellent error handling and retry capabilities

### MCP Protocol
- **[github.com/mark3labs/mcp-go](https://github.com/mark3labs/mcp-go)** `v0.4.0`
  - **Purpose**: Model Context Protocol implementation in Go
  - **Features**: Complete MCP protocol support, STDIO transport, tool registration
  - **Why chosen**: Official Go implementation, saves development time, maintains protocol compatibility

### Configuration & Environment
- **[github.com/joho/godotenv](https://github.com/joho/godotenv)** `v1.5.1`
  - **Purpose**: Load environment variables from .env files
  - **Features**: Simple .env file parsing, development environment support
  - **Why chosen**: Standard solution for Go projects, simple and reliable

### Logging
- **[github.com/rs/zerolog](https://github.com/rs/zerolog)** `v1.31.0`
  - **Purpose**: Structured, high-performance logging
  - **Features**: Zero allocation logging, JSON output, structured fields, multiple output formats
  - **Why chosen**: Fastest Go logger, zero allocation design, excellent structured logging support

### Rate Limiting
- **[golang.org/x/time](https://golang.org/x/time)** `v0.5.0`
  - **Purpose**: Rate limiting for API calls using token bucket algorithm
  - **Features**: Token bucket rate limiter, context support
  - **Why chosen**: Official Go extended package, proven algorithm, context-aware

## Build Tools

### Build Automation
- **[github.com/magefile/mage](https://github.com/magefile/mage)** `v1.15.0`
  - **Purpose**: Go-based build tool (replacement for Makefiles)
  - **Features**: Go-native build scripts, cross-platform, IDE integration
  - **Why chosen**: Native Go solution, better than Makefiles for Go projects, type safety

## Development Dependencies (Planned/Future)

### Testing
- **github.com/stretchr/testify** - Testing assertions and test suites
- **github.com/golang/mock** - Mock generation for interfaces

### Validation
- **github.com/go-playground/validator** - Struct validation for requests

### Markdown Processing
- **github.com/gomarkdown/markdown** - Markdown parsing and conversion

## Dependency Analysis

### Security Considerations
- All dependencies are from reputable sources
- Regular security updates through `go mod tidy`
- No known security vulnerabilities in current versions

### Performance Impact
- **zerolog**: Zero allocation logging - minimal performance impact
- **resty**: Efficient HTTP client with connection pooling
- **golang.org/x/time**: Lightweight rate limiting
- **mcp-go**: Protocol handling with minimal overhead

### License Compatibility
All dependencies use permissive licenses compatible with commercial use:
- MIT License: resty, godotenv, zerolog, mage
- BSD License: golang.org/x/time, mcp-go

### Maintenance Status
All dependencies are actively maintained:
- Regular updates and bug fixes
- Active community support
- Long-term stability expected

## Version Management

### Go Version
- **Required**: Go 1.21+
- **Tested with**: Go 1.23.2
- **Modules**: Full Go modules support

### Dependency Updates
- Use `go mod tidy` to maintain clean dependencies
- Use `go list -m -u all` to check for updates
- Regular security updates recommended

### Build Reproducibility
- `go.sum` file ensures reproducible builds
- All dependencies are pinned to specific versions
- Cross-platform compatibility maintained

## Architecture Impact

### Layered Design
- **Transport Layer**: mcp-go handles protocol communication
- **HTTP Layer**: resty manages Nuclino API calls
- **Application Layer**: Custom business logic with minimal dependencies
- **Infrastructure Layer**: zerolog for observability, mage for builds

### Dependency Injection
- Interfaces used throughout for testability
- Dependencies injected at startup
- Easy to mock for testing

### Error Handling
- Structured error handling with context
- API errors properly categorized
- Logging integrated with error flows

## Future Considerations

### Potential Additions
- **OpenTelemetry**: Distributed tracing and metrics
- **Prometheus client**: Metrics collection
- **Validator**: Request validation
- **CLI framework**: Enhanced command-line interface

### Migration Path
- Current dependencies chosen for long-term stability
- Easy to add new dependencies without breaking changes
- Minimal vendor lock-in risk
# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is a Model Context Protocol (MCP) server implementation for Nuclino API written in Go. It enables integration with Claude Desktop and other MCP clients to interact with Nuclino workspaces, collections, items, and files.

## Development Commands

This project uses [Mage](https://magefile.org/) instead of Makefiles for build automation. Install Mage first if you haven't:
```bash
go install github.com/magefile/mage@latest
```

### Building and Running
- `mage build` - Build the binary to `bin/nuclino-mcp-server`
- `mage run` - Build and run the server
- `mage dev` - Build and run the server in debug mode with verbose logging
- `mage install` - Download and organize Go dependencies

### Testing and Quality
- `mage test` - Run all tests
- `mage testcoverage` - Run tests with coverage report (generates coverage.html)
- `mage benchmark` - Run benchmark tests
- `mage lint` - Run golangci-lint (installs if not present)
- `mage fmt` - Format all Go code
- `mage security` - Run gosec security scanner

### Build Variants
- `mage buildall` - Cross-compile for multiple platforms (Linux, macOS, Windows)
- `mage docker` - Build Docker image

### Development Tools
- `mage installtools` - Install development tools (mockgen, etc.)
- `mage generatemocks` - Generate test mocks from interfaces

### Convenience Commands
- `mage ci` - Run all CI tasks (install, fmt, lint, test, build)
- `mage all` - Run all development tasks (install, fmt, lint, testcoverage, build)
- `mage clean` - Remove build artifacts
- `mage checkgo` - Check Go version

### Available Targets
Run `mage -l` to see all available targets with descriptions.

## Architecture

### Core Components
- **`cmd/server/main.go`** - Application entry point with CLI flags and environment setup
- **`internal/mcp/`** - MCP protocol implementation (JSON-RPC over STDIO)
- **`internal/nuclino/`** - Nuclino API client with rate limiting and retry logic
- **`internal/tools/`** - MCP tool implementations for Nuclino operations

### MCP Tools Available
1. **Items**: `nuclino_get_item`, `nuclino_search_items`, `nuclino_create_item`, `nuclino_update_item`, `nuclino_delete_item`, `nuclino_move_item`
2. **Workspaces**: `nuclino_list_workspaces`, `nuclino_get_workspace`, `nuclino_create_workspace`, `nuclino_update_workspace`, `nuclino_delete_workspace`
3. **Collections**: `nuclino_list_collections`, `nuclino_get_collection`, `nuclino_create_collection`, `nuclino_update_collection`, `nuclino_delete_collection`
4. **Users/Teams**: `nuclino_get_current_user`, `nuclino_get_user`, `nuclino_list_teams`, `nuclino_get_team`
5. **Files**: `nuclino_list_files`, `nuclino_get_file`

### Key Libraries Used
- **github.com/go-resty/resty/v2** - HTTP client with automatic retry
- **github.com/rs/zerolog** - Structured logging
- **golang.org/x/time/rate** - Rate limiting for API calls
- **github.com/go-playground/validator/v10** - Request validation

## Configuration

### Environment Variables
Required:
- `NUCLINO_API_KEY` - Your Nuclino API key

Optional (see `.env.example`):
- `LOG_LEVEL` - Log level (debug, info, warn, error)
- `DEBUG` - Enable debug mode
- `RATE_LIMIT_RPS` - Requests per second limit
- `HTTP_TIMEOUT` - HTTP client timeout

### Running the Server
1. Copy `.env.example` to `.env`
2. Set your `NUCLINO_API_KEY`
3. Run with `mage dev` for development or `mage run` for production

## Development Guidelines

### Adding New Tools
1. Create tool struct in appropriate file in `internal/tools/`
2. Implement the `Tool` interface: `Name()`, `Description()`, `InputSchema()`, `Execute()`
3. Register the tool in `internal/tools/registry.go`
4. Use `JSONSchema()`, `StringProperty()`, `IntProperty()` helpers for schema definition
5. Use `FormatResult()` and `FormatError()` for consistent responses

### Error Handling
- Use typed errors from `internal/nuclino/errors.go`
- Check for specific API errors: `IsNotFound()`, `IsUnauthorized()`, `IsRateLimited()`
- Always format errors properly with `FormatError()` for MCP responses

### Testing
- Unit tests should go in `*_test.go` files alongside source
- Integration tests go in `tests/integration/`
- Generate mocks with `mage generatemocks`
- Target 80%+ test coverage

### Code Style
- Follow Go conventions: `go fmt`, `go vet`
- Use interfaces for testability
- Implement proper context handling for all API calls
- Use structured logging with zerolog

## MCP Protocol Notes

- Server communicates via JSON-RPC 2.0 over STDIN/STDOUT
- Implements required methods: `initialize`, `tools/list`, `tools/call`
- All tools return JSON-formatted results wrapped in MCP `ToolsCallResult`
- Rate limiting is applied transparently to all Nuclino API calls

## Common Issues

1. **Authentication errors** - Verify NUCLINO_API_KEY is set correctly
2. **Rate limiting** - Server automatically handles rate limits with exponential backoff
3. **Network timeouts** - Adjust HTTP_TIMEOUT environment variable if needed
4. **Build issues** - Ensure Go 1.21+ is installed, run `mage install`
# Nuclino MCP Server

Enterprise-grade Model Context Protocol (MCP) server for Nuclino API integration with Claude Desktop. Features advanced rate limiting, intelligent caching, and comprehensive error handling.

## ⚡ Quick Start

### Prerequisites
- Go 1.21+
- [Mage](https://magefile.org/) - `go install github.com/magefile/mage@latest`
- Nuclino API key - [Get yours here](https://help.nuclino.com/d3a29686-api)

### Installation

```bash
git clone https://github.com/lukasz/nuclino-mcp.git
cd nuclino-mcp
mage install  # Install dependencies
cp .env.example .env  # Add your NUCLINO_API_KEY
mage build
```

### Claude Desktop Setup

Add to `claude_desktop_config.json`:

```json
{
  "mcpServers": {
    "nuclino": {
      "command": "/full/path/to/nuclino-mcp/bin/nuclino-mcp-server",
      "env": {
        "NUCLINO_API_KEY": "your_api_key_here"
      }
    }
  }
}
```

**📖 [Complete Setup Guide](docs/CLAUDE_DESKTOP_SETUP.md)**

## 🛠 Features

### ✅ 29+ MCP Tools
- **Items:** CRUD operations, search, bulk management
- **Workspaces:** Overview, analysis, content search  
- **Collections:** Organization, bulk operations, statistics
- **Users/Teams:** User management and team operations
- **Files:** File listing and metadata

### 🚀 Enterprise Features
- **Rate Limiting:** Circuit breaker pattern with adaptive control
- **Intelligent Caching:** TTL-based with LRU eviction
- **Error Handling:** Categorized errors with automatic retries
- **Monitoring:** Performance metrics and health checks
- **Performance:** Stress tested, benchmarked, memory-bounded

### 🎯 Usage Examples

```
Claude, list my Nuclino workspaces

Search for "API documentation" in my Nuclino workspace

Create a new item titled "Meeting Notes" in collection "xyz789"

Give me a comprehensive overview of workspace "workspace-123"

Analyze collection "docs-456" and suggest organization improvements
```

## 📚 Documentation

| Guide | Description |
|-------|-------------|
| **[Claude Desktop Setup](docs/CLAUDE_DESKTOP_SETUP.md)** | Complete integration guide |
| **[Tools Reference](docs/TOOLS_REFERENCE.md)** | All 29+ tools with examples |
| **[Development Guide](docs/DEVELOPMENT_GUIDE.md)** | Extending and building |
| **[Troubleshooting](docs/TROUBLESHOOTING.md)** | Common issues and solutions |
| **[Enhanced Features](ENHANCED_FEATURES.md)** | Advanced capabilities |

## 🧪 Development

### Quick Commands
```bash
mage build        # Build binary
mage test         # Run all tests
mage dev          # Run with debug logging
mage ci           # Full CI pipeline
```

### Testing
- **Unit Tests:** 29+ tools with comprehensive mocks
- **Integration Tests:** Multi-tool workflows
- **Performance Tests:** Stress testing and benchmarks
- **Error Handling:** Edge cases and failure scenarios

## 🔧 Configuration

```bash
# Required
NUCLINO_API_KEY=your_nuclino_api_key

# Optional (with defaults)
LOG_LEVEL=info           # debug, info, warn, error
RATE_LIMIT_RPS=10        # API requests per second  
HTTP_TIMEOUT=30s         # HTTP client timeout
CACHE_TTL=300s          # Cache expiration time
CACHE_SIZE=1000         # Maximum cache entries
```

## 🐛 Troubleshooting

### Common Issues

| Issue | Solution |
|-------|----------|
| "NUCLINO_API_KEY required" | Add API key to .env or Claude config |
| "Server failed to start" | Check binary path and permissions |
| "Tool call failed" | Verify API key and connectivity |

**📖 [Full Troubleshooting Guide](docs/TROUBLESHOOTING.md)**

## 📊 Project Status

**Phase 4 Complete**: Enterprise-Grade Features

- ✅ MCP Server with official `mcp-go` library
- ✅ 29+ Tools with complete Nuclino API coverage
- ✅ Advanced rate limiting with circuit breaker pattern
- ✅ Intelligent caching with TTL and LRU eviction
- ✅ Comprehensive error handling with automatic retries
- ✅ Performance monitoring and health checks
- ✅ Extensive testing (unit + integration + performance)
- ✅ Cross-platform builds and CI/CD automation

## 🤝 Contributing

1. Fork the repository
2. Create feature branch: `git checkout -b feature-name`  
3. Add tests for new features
4. Run: `mage ci` (format, lint, test, build)
5. Submit pull request

## 📄 License

MIT License - see [LICENSE](LICENSE) file for details.

---

**Built with ❤️ using Go and the official mcp-go library**
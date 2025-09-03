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
      "command": "/full/path/to/nuclino-mcp/scripts/mcp-wrapper.sh",
      "args": ["/full/path/to/nuclino-mcp/bin/nuclino-mcp-server"],
      "env": {
        "NUCLINO_API_KEY": "your_api_key_here"
      }
    }
  }
}
```

**💡 Important:** Use the wrapper script to prevent JSON-RPC protocol issues with Claude Desktop.

**📖 [Complete Setup Guide](docs/CLAUDE_DESKTOP_SETUP.md)**

## 🛠 Features

### ✅ 18 Working MCP Tools
- **Items:** Create, read, update, delete, search, list
- **Workspaces:** List, get details, overview, content search  
- **Users/Teams:** User info, team management
- **Files:** File listing and metadata

**📊 API Status:** 87% of core functionality working (based on official API testing)

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

Create a new item titled "Meeting Notes" with workspace_id "abc123"

Give me a comprehensive overview of workspace "workspace-123"

Update the item "item-456" with new content

Delete the old draft item "draft-789"
```

## 📚 Documentation

| Guide | Description |
|-------|-------------|
| **[Claude Desktop Setup](docs/CLAUDE_DESKTOP_SETUP.md)** | Complete integration guide |
| **[Tools Reference](docs/TOOLS_REFERENCE.md)** | All 18 working tools with examples |
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
- **Unit Tests:** 18 working tools with comprehensive mocks
- **Integration Tests:** Multi-tool workflows
- **API Testing:** Real endpoint verification against production API
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

**Phase 5 Complete**: Production Ready with Real API Integration

- ✅ MCP Server with official `mcp-go` library
- ✅ 18 Working tools with verified Nuclino API endpoints
- ✅ Real API testing against production Nuclino API
- ✅ Advanced rate limiting with circuit breaker pattern
- ✅ Intelligent caching with TTL and LRU eviction
- ✅ Comprehensive error handling with automatic retries
- ✅ Performance monitoring and health checks
- ✅ Extensive testing (unit + integration + performance + API)
- ✅ Cross-platform builds and CI/CD automation
- ✅ Complete documentation with working examples

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
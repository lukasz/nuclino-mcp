# Claude Desktop Setup Guide

Complete guide for integrating Nuclino MCP Server with Claude Desktop.

## Prerequisites

- Claude Desktop application installed
- Nuclino MCP Server built and available
- Valid Nuclino API key

## Step-by-Step Setup

### 1. Get Your Nuclino API Key

1. Log into your Nuclino account
2. Go to **Settings** → **API** 
3. Generate a new API key
4. Copy the key - you'll need it for configuration

### 2. Locate Claude Desktop Config File

The configuration file location depends on your operating system:

**macOS:**
```bash
~/Library/Application Support/Claude/claude_desktop_config.json
```

**Windows:**
```bash
%APPDATA%\Claude\claude_desktop_config.json
```

**Linux:**
```bash
~/.config/Claude/claude_desktop_config.json
```

### 3. Configure Claude Desktop

Create or edit the `claude_desktop_config.json` file:

```json
{
  "mcpServers": {
    "nuclino": {
      "command": "/full/path/to/nuclino-mcp/scripts/mcp-wrapper.sh",
      "args": ["/full/path/to/nuclino-mcp/bin/nuclino-mcp-server"],
      "env": {
        "NUCLINO_API_KEY": "your_api_key_here",
        "LOG_LEVEL": "info",
        "RATE_LIMIT_RPS": "10"
      }
    }
  }
}
```

**Important:** Use the wrapper script to avoid JSON-RPC protocol issues. Alternatively, you can use the server directly:

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

*Note: The wrapper fixes protocol compatibility issues with Claude Desktop. If you experience connection problems, always try the wrapper first.*

**Important Notes:**
- Use **full absolute path** to the binary
- Replace `your_api_key_here` with your actual API key
- Ensure the binary has execute permissions: `chmod +x bin/nuclino-mcp-server`

### 4. Verify Installation

1. **Restart Claude Desktop** completely
2. Open a new conversation
3. Test with simple commands:

```
Claude, list my Nuclino workspaces
```

```
Claude, search for "documentation" in my Nuclino workspace
```

## Advanced Configuration

### Environment Variables

You can configure additional settings in the `env` section:

```json
{
  "mcpServers": {
    "nuclino": {
      "command": "/path/to/nuclino-mcp-server",
      "env": {
        "NUCLINO_API_KEY": "your_api_key",
        "LOG_LEVEL": "debug",
        "RATE_LIMIT_RPS": "5",
        "HTTP_TIMEOUT": "30s",
        "CACHE_TTL": "300s",
        "CACHE_SIZE": "1000"
      }
    }
  }
}
```

### Configuration Options

| Variable | Default | Description |
|----------|---------|-------------|
| `NUCLINO_API_KEY` | Required | Your Nuclino API key |
| `LOG_LEVEL` | `info` | Logging level: debug, info, warn, error |
| `RATE_LIMIT_RPS` | `10` | API requests per second |
| `HTTP_TIMEOUT` | `30s` | HTTP request timeout |
| `CACHE_TTL` | `300s` | Cache entry time-to-live |
| `CACHE_SIZE` | `1000` | Maximum cache entries |

## Usage Examples

### Basic Operations

**List workspaces:**
```
Show me all my Nuclino workspaces
```

**Search content:**
```
Search for "meeting notes" in my Nuclino workspace
```

**Get specific item:**
```
Get the content of Nuclino item with ID "abc123"
```

### Advanced Operations

**Create new content:**
```
Create a new Nuclino item titled "Project Overview" in collection "xyz789" with this content:
# Project Overview
## Goals
- Goal 1
- Goal 2
## Timeline
Q1 2024
```

**Workspace analysis:**
```
Give me a comprehensive overview of workspace "workspace-123" including collections and recent activity
```

**Collection organization:**
```
Analyze collection "docs-456" and suggest organization improvements
```

## Troubleshooting

### Server Not Starting

1. **Check binary path:**
   ```bash
   /path/to/nuclino-mcp-server --version
   ```

2. **Verify permissions:**
   ```bash
   chmod +x /path/to/nuclino-mcp-server
   ```

3. **Test standalone:**
   ```bash
   NUCLINO_API_KEY=your_key /path/to/nuclino-mcp-server
   ```

### API Key Issues

1. **Verify API key format:**
   - Should be a long alphanumeric string
   - No spaces or special characters

2. **Test API access:**
   ```bash
   curl -H "Authorization: Bearer your_api_key" https://api.nuclino.com/v0/workspaces
   ```

### Connection Problems

1. **Check Claude Desktop logs:**
   - macOS: `~/Library/Logs/Claude/`
   - Windows: `%APPDATA%\Claude\logs\`

2. **Enable debug logging:**
   ```json
   "env": {
     "NUCLINO_API_KEY": "your_key",
     "LOG_LEVEL": "debug"
   }
   ```

3. **Restart both Claude Desktop and server:**
   - Quit Claude Desktop completely
   - Wait 10 seconds
   - Restart Claude Desktop

### Common Error Messages

**"Server failed to start"**
- Check binary path and permissions
- Verify API key is set correctly

**"Tool call failed"**
- API key may be invalid or expired
- Network connectivity issues
- Item/workspace ID doesn't exist

**"Rate limit exceeded"**
- Server automatically handles rate limiting
- Reduce `RATE_LIMIT_RPS` if needed

## Advanced Features

### Enterprise Features

The server includes enterprise-grade features:

- **Rate Limiting:** Circuit breaker pattern with adaptive adjustment
- **Intelligent Caching:** TTL-based with LRU eviction
- **Error Handling:** Automatic retries with exponential backoff
- **Monitoring:** Performance metrics and health checks

### Performance Optimization

For high-volume usage, adjust these settings:

```json
"env": {
  "RATE_LIMIT_RPS": "15",
  "CACHE_SIZE": "2000",
  "CACHE_TTL": "600s"
}
```

### Debugging

Enable detailed logging:

```json
"env": {
  "LOG_LEVEL": "debug",
  "DEBUG": "true"
}
```

This will provide detailed logs for troubleshooting connection and API issues.

## Getting Help

- Check the main [README.md](../README.md) for general information
- Review [ENHANCED_FEATURES.md](../ENHANCED_FEATURES.md) for advanced capabilities
- Test connectivity with the standalone test: `go run test_mcp_server.go`
# Troubleshooting Guide

Comprehensive troubleshooting guide for common issues with the Nuclino MCP Server.

## Quick Diagnostics

### Health Check Commands

```bash
# Test server binary
./bin/nuclino-mcp-server --version

# Test API connectivity
go run test_mcp_server.go

# Run with debug logging
LOG_LEVEL=debug ./bin/nuclino-mcp-server

# Check configuration
mage dev
```

### Environment Verification

```bash
# Check Go version (requires 1.21+)
go version

# Verify dependencies
mage install

# Test build
mage build
```

## Common Issues

### 1. Server Won't Start

**Symptoms:**
- Server binary doesn't run
- "Permission denied" errors
- "Command not found" errors

**Solutions:**

**Check binary permissions:**
```bash
chmod +x /path/to/nuclino-mcp-server
ls -la /path/to/nuclino-mcp-server
```

**Verify binary path:**
```bash
which nuclino-mcp-server
/full/path/to/nuclino-mcp-server --version
```

**Test standalone execution:**
```bash
NUCLINO_API_KEY=your_key /path/to/nuclino-mcp-server
```

**Check for missing dependencies:**
```bash
ldd /path/to/nuclino-mcp-server  # Linux
otool -L /path/to/nuclino-mcp-server  # macOS
```

### 2. API Key Issues

**Symptoms:**
- "NUCLINO_API_KEY required" error
- "Authentication failed" responses
- "Invalid API key" messages

**Solutions:**

**Verify API key format:**
- Should be a long alphanumeric string
- No spaces or special characters
- Typically 40+ characters long

**Test API key manually:**
```bash
curl -H "Authorization: Bearer your_api_key" https://api.nuclino.com/v0/workspaces
```

**Check environment variable:**
```bash
echo $NUCLINO_API_KEY
env | grep NUCLINO
```

**Verify .env file:**
```bash
cat .env
# Should contain: NUCLINO_API_KEY=your_actual_key
```

### 3. Claude Desktop Integration Issues

**Symptoms:**
- Server not appearing in Claude Desktop
- "Server failed to start" in Claude
- Tools not available in conversations

**Solutions:**

**Verify Claude Desktop config path:**
- macOS: `~/Library/Application Support/Claude/claude_desktop_config.json`
- Windows: `%APPDATA%\Claude\claude_desktop_config.json`
- Linux: `~/.config/Claude/claude_desktop_config.json`

**Check config file syntax:**
```json
{
  "mcpServers": {
    "nuclino": {
      "command": "/full/absolute/path/to/nuclino-mcp-server",
      "env": {
        "NUCLINO_API_KEY": "your_api_key_here"
      }
    }
  }
}
```

**Common config mistakes:**
- Using relative paths instead of absolute paths
- Missing environment variables
- JSON syntax errors (missing commas, quotes)
- Incorrect binary path

**Test config:**
```bash
# Validate JSON syntax
cat claude_desktop_config.json | jq .

# Test binary from config
/path/from/config --version
```

**Restart process:**
1. Quit Claude Desktop completely
2. Wait 10 seconds
3. Restart Claude Desktop
4. Test with: "List my Nuclino workspaces"

### 4. Network and Connectivity Issues

**Symptoms:**
- "Network error" messages
- Timeouts during API calls
- Intermittent connection failures

**Solutions:**

**Check internet connectivity:**
```bash
ping api.nuclino.com
curl -I https://api.nuclino.com/v0/workspaces
```

**Test DNS resolution:**
```bash
nslookup api.nuclino.com
dig api.nuclino.com
```

**Check firewall/proxy settings:**
```bash
# Set proxy if needed
export HTTP_PROXY=http://proxy.company.com:8080
export HTTPS_PROXY=https://proxy.company.com:8080
```

**Verify TLS/SSL:**
```bash
openssl s_client -connect api.nuclino.com:443
```

**Adjust timeout settings:**
```json
"env": {
  "NUCLINO_API_KEY": "your_key",
  "HTTP_TIMEOUT": "60s"
}
```

### 5. Rate Limiting Issues

**Symptoms:**
- "Rate limit exceeded" errors
- "Too many requests" responses
- Slow response times

**Solutions:**

**Check rate limit settings:**
```json
"env": {
  "RATE_LIMIT_RPS": "5",
  "CACHE_TTL": "600s"
}
```

**Monitor rate limiting:**
```bash
# Enable debug logging to see rate limit stats
LOG_LEVEL=debug ./nuclino-mcp-server
```

**Adjust rate limiting:**
- Reduce `RATE_LIMIT_RPS` (default: 10)
- Increase `CACHE_TTL` (default: 300s)
- Enable adaptive rate limiting

### 6. Memory and Performance Issues

**Symptoms:**
- High memory usage
- Slow response times
- Server crashes or freezes

**Solutions:**

**Monitor resource usage:**
```bash
# Check memory usage
ps aux | grep nuclino-mcp-server

# Monitor during operation
top -p $(pgrep nuclino-mcp-server)
```

**Optimize cache settings:**
```json
"env": {
  "CACHE_SIZE": "500",
  "CACHE_TTL": "300s"
}
```

**Enable performance monitoring:**
```bash
# Run with profiling
go build -tags profile
./nuclino-mcp-server -cpuprofile=cpu.prof -memprofile=mem.prof
```

## Advanced Debugging

### Debug Logging

Enable comprehensive logging:

```json
"env": {
  "LOG_LEVEL": "debug",
  "DEBUG": "true"
}
```

Log levels available:
- `debug`: Detailed debugging information
- `info`: General operational messages
- `warn`: Warning conditions
- `error`: Error conditions only

### Request Tracing

Enable request tracing for API calls:

```bash
# Set debug mode
DEBUG=true LOG_LEVEL=debug ./nuclino-mcp-server
```

This will show:
- HTTP request/response details
- Rate limiting decisions
- Cache hit/miss statistics
- Circuit breaker state changes

### Performance Profiling

For performance issues:

```bash
# CPU profiling
go tool pprof http://localhost:6060/debug/pprof/profile?seconds=30

# Memory profiling  
go tool pprof http://localhost:6060/debug/pprof/heap

# Goroutine analysis
go tool pprof http://localhost:6060/debug/pprof/goroutine
```

### Health Monitoring

Check server health:

```bash
# Test with health check endpoint (if enabled)
curl http://localhost:8080/health

# Check metrics (if enabled)
curl http://localhost:8080/metrics
```

## Error Messages Reference

### Common Error Patterns

**"NUCLINO_API_KEY required"**
- Missing or empty API key
- Check environment variables and .env file

**"Tool call failed: unauthorized"**
- Invalid or expired API key  
- API key lacks required permissions

**"Tool call failed: not found"**
- Invalid workspace, collection, or item ID
- Resource may have been deleted

**"Tool call failed: rate limit exceeded"**
- Too many requests in short timeframe
- Server will automatically retry

**"Tool call failed: network error"**
- Connectivity issues
- DNS resolution problems
- Firewall blocking requests

**"Tool call failed: timeout"**
- Request took too long
- Increase HTTP_TIMEOUT setting

**"Server failed to start"**
- Binary path incorrect in Claude config
- Missing permissions on binary
- Environment variables not accessible

### HTTP Status Code Meanings

- `400 Bad Request`: Invalid parameters or malformed request
- `401 Unauthorized`: Invalid or missing API key
- `403 Forbidden`: API key lacks required permissions
- `404 Not Found`: Resource doesn't exist
- `429 Too Many Requests`: Rate limit exceeded
- `500 Internal Server Error`: Server-side error
- `502/503/504`: Service temporarily unavailable

## Platform-Specific Issues

### macOS

**Security warnings:**
```bash
# Allow execution of unsigned binary
sudo xattr -r -d com.apple.quarantine /path/to/nuclino-mcp-server
```

**Path issues:**
- Use full paths in Claude config
- Avoid spaces in directory names
- Check PATH environment in Claude Desktop

### Windows

**PowerShell execution:**
```powershell
# Set execution policy if needed
Set-ExecutionPolicy -ExecutionPolicy RemoteSigned -Scope CurrentUser

# Check binary
.\nuclino-mcp-server.exe --version
```

**Path separators:**
- Use forward slashes or escaped backslashes in config
- Ensure binary has .exe extension

### Linux

**Dependencies:**
```bash
# Install required libraries
sudo apt-get update
sudo apt-get install ca-certificates

# For older systems
sudo apt-get install libc6
```

**Permissions:**
```bash
# Ensure execute permission
chmod +x nuclino-mcp-server

# Check library dependencies
ldd nuclino-mcp-server
```

## Getting Help

### Information to Gather

When reporting issues, include:

1. **Environment:**
   - Operating system and version
   - Go version (`go version`)
   - Binary version (`./nuclino-mcp-server --version`)

2. **Configuration:**
   - Claude Desktop config (redact API key)
   - Environment variables (redact sensitive values)

3. **Error details:**
   - Complete error messages
   - Debug logs if possible
   - Steps to reproduce

4. **Network info:**
   - Internet connectivity test results
   - Proxy/firewall configuration

### Testing Commands

Run these to gather diagnostic information:

```bash
# System info
uname -a
go version

# Binary info
./nuclino-mcp-server --version
ldd ./nuclino-mcp-server  # Linux
otool -L ./nuclino-mcp-server  # macOS

# Network tests
ping api.nuclino.com
curl -I https://api.nuclino.com/v0/workspaces

# API test (redact key in logs)
NUCLINO_API_KEY=your_key go run test_mcp_server.go

# Debug run
LOG_LEVEL=debug ./nuclino-mcp-server
```

### Support Channels

- Check the main [README.md](../README.md) for general information
- Review [CLAUDE_DESKTOP_SETUP.md](CLAUDE_DESKTOP_SETUP.md) for integration help
- See [DEVELOPMENT_GUIDE.md](DEVELOPMENT_GUIDE.md) for technical details
- Test with the provided utilities in the repository
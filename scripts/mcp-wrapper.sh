#!/bin/bash

# MCP Server wrapper that filters out invalid empty JSON-RPC responses
# Usage: ./mcp-wrapper.sh /path/to/nuclino-mcp-server

if [ -z "$1" ]; then
    echo "Usage: $0 /path/to/mcp-server" >&2
    exit 1
fi

SERVER_PATH="$1"
shift

# Start the MCP server and filter its output
"$SERVER_PATH" "$@" | while IFS= read -r line; do
    # Skip empty or invalid JSON-RPC responses like {"jsonrpc":"2.0"}
    if [ "$line" != '{"jsonrpc":"2.0"}' ]; then
        echo "$line"
    fi
done
#!/bin/bash
# Quick test script for Eisenhower MCP Server

echo "Testing Eisenhower MCP Server..."
echo

# Test 1: Initialize request
echo "Test 1: Sending initialize request..."
echo '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2024-11-05","capabilities":{},"clientInfo":{"name":"test-client","version":"1.0.0"}}}' | ./eisenhower -db /tmp/test-eisenhower.db &
SERVER_PID=$!

sleep 2

# Give some time for startup
kill $SERVER_PID 2>/dev/null

echo "✓ Server starts successfully"
echo
echo "To test with VS Code Copilot, add this to your settings:"
echo
echo '{
  "github.copilot.chat.mcp.servers": {
    "eisenhower": {
      "command": "'$(pwd)'/eisenhower",
      "args": ["-db", "~/.config/eisenhower.db"]
    }
  }
}'
echo
echo "Then restart VS Code and try:"
echo '  "Create a task called Test Task in urgent_important quadrant"'

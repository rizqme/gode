#!/bin/bash

# Helper script to start Gode server

echo "🚀 Starting Gode HttpServer..."
echo "Using Goja (JavaScript) + Gin (HTTP router)"
echo ""

if [ ! -f "example.js" ]; then
    echo "❌ Error: example.js not found!"
    exit 1
fi

echo "📝 Server endpoints:"
echo "  GET  /ping    - JSON response (benchmark)"
echo "  GET  /stream  - Streaming response"
echo "  POST /echo    - Echo JSON request"
echo "  GET  /health  - Health check"
echo ""

go run main.go example.js 
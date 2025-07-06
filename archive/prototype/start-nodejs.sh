#!/bin/bash

# Helper script to start Node.js server

echo "🚀 Starting Node.js Express server..."
echo "Using Express.js HTTP framework"
echo ""

if [ ! -f "baseline/app.js" ]; then
    echo "❌ Error: baseline/app.js not found!"
    exit 1
fi

echo "📝 Server endpoints:"
echo "  GET  /ping    - JSON response (benchmark)"
echo "  GET  /stream  - Streaming response"
echo "  POST /echo    - Echo JSON request"
echo "  GET  /health  - Health check"
echo ""

cd baseline && npm install --silent && node app.js 
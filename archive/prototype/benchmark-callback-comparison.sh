#!/bin/bash

# 🚨 Go-Powered Callbacks vs Node.js setTimeout Benchmark
# This compares Go goroutines vs Node.js setTimeout for external API simulation

echo "🚨 GO-POWERED CALLBACKS VS NODE.JS SETTIMEOUT BENCHMARK"
echo "========================================================="
echo "🔥 Go Version: External API calls handled by Go goroutines + callbacks"
echo "🔥 Node.js Version: External API calls handled by setTimeout + callbacks"
echo

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Kill any existing processes
echo "🧹 Cleaning up existing processes..."
lsof -ti:8080 | xargs -r kill -9 2>/dev/null || true
sleep 2

# Test function to wait for server
wait_for_server() {
    echo "⏳ Waiting for server to start..."
    for i in {1..30}; do
        if curl -s http://localhost:8080/health > /dev/null 2>&1; then
            echo "✅ Server is ready!"
            return 0
        fi
        sleep 1
    done
    echo "❌ Server failed to start"
    return 1
}

# Test endpoint performance
test_endpoint() {
    local name=$1
    local endpoint=$2
    local description=$3
    
    echo -e "${BLUE}Testing ${name} - ${description}${NC}"
    echo "Endpoint: ${endpoint}"
    echo "📊 Performance (4 threads, 10 connections, 10 seconds):"
    wrk -t4 -c10 -d10s --latency http://localhost:8080${endpoint} | grep -E "(Requests/sec|Latency|requests in)"
    echo
}

# Test response time for single requests
test_response_time() {
    local name=$1
    local endpoint=$2
    local description=$3
    
    echo -e "${BLUE}Testing ${name} - ${description} Response Time${NC}"
    
    echo "📊 Single request timings (5 samples):"
    for i in {1..5}; do
        start_time=$(date +%s%N)
        curl -s "http://localhost:8080${endpoint}" > /dev/null 2>&1
        end_time=$(date +%s%N)
        duration=$(( (end_time - start_time) / 1000000 ))  # Convert to ms
        echo "  Request #${i}: ${duration}ms"
    done
    echo
}

# Test concurrent request handling
test_concurrent_load() {
    local name=$1
    local endpoint=$2
    
    echo -e "${BLUE}Testing ${name} - Concurrent Load Handling${NC}"
    
    echo "📊 Light concurrent load (2 threads, 5 connections):"
    wrk -t2 -c5 -d5s --latency http://localhost:8080${endpoint} | grep -E "(Requests/sec|Latency)"
    
    echo
    echo "📊 Medium concurrent load (4 threads, 20 connections):"
    wrk -t4 -c20 -d5s --latency http://localhost:8080${endpoint} | grep -E "(Requests/sec|Latency)"
    
    echo
    echo "📊 Heavy concurrent load (8 threads, 50 connections):"
    wrk -t8 -c50 -d5s --latency http://localhost:8080${endpoint} | grep -E "(Requests/sec|Latency)"
    
    echo
}

echo "🚀 TESTING GODE (Go-Powered Callbacks)"
echo "======================================="

# Start Gode server
echo "Starting Gode server with Go-powered callbacks..."
timeout 120 go run main.go callback-test.js > /dev/null 2>&1 &
GODE_PID=$!

if wait_for_server; then
    test_endpoint "GODE" "/test-go-api" "Single Go goroutine API call"
    test_endpoint "GODE" "/test-go-batch" "3 concurrent Go goroutine API calls"
    test_endpoint "GODE" "/test-user-profile/test123" "5 concurrent Go goroutine API calls"
    
    test_response_time "GODE" "/test-go-api" "Single API call"
    test_response_time "GODE" "/test-user-profile/test123" "5 concurrent API calls"
    
    test_concurrent_load "GODE" "/test-user-profile/test123"
    
    echo -e "${GREEN}✅ Gode (Go-powered callbacks) testing completed${NC}"
else
    echo -e "${RED}❌ Gode server failed to start${NC}"
fi

# Kill Gode server
kill $GODE_PID 2>/dev/null || true
sleep 3

echo
echo "🚀 TESTING NODE.JS (setTimeout Callbacks)"
echo "=========================================="

# Start Node.js server
echo "Starting Node.js server with setTimeout callbacks..."
cd baseline
timeout 120 node callback-test.js > /dev/null 2>&1 &
NODE_PID=$!
cd ..

if wait_for_server; then
    test_endpoint "NODE.JS" "/test-go-api" "Single setTimeout API call"
    test_endpoint "NODE.JS" "/test-go-batch" "3 concurrent setTimeout API calls"
    test_endpoint "NODE.JS" "/test-user-profile/test123" "5 concurrent setTimeout API calls"
    
    test_response_time "NODE.JS" "/test-go-api" "Single API call"
    test_response_time "NODE.JS" "/test-user-profile/test123" "5 concurrent API calls"
    
    test_concurrent_load "NODE.JS" "/test-user-profile/test123"
    
    echo -e "${GREEN}✅ Node.js (setTimeout callbacks) testing completed${NC}"
else
    echo -e "${RED}❌ Node.js server failed to start${NC}"
fi

# Kill Node.js server
kill $NODE_PID 2>/dev/null || true

echo
echo "🏁 BENCHMARK COMPLETE"
echo "===================="
echo -e "${YELLOW}KEY DIFFERENCES:${NC}"
echo "• Go Version: External API calls handled by Go goroutines"
echo "• Node.js Version: External API calls handled by JavaScript setTimeout"
echo
echo -e "${YELLOW}WHAT TO COMPARE:${NC}"
echo "1. Response times for endpoints with multiple API calls"
echo "2. Throughput under concurrent load"
echo "3. Latency consistency (lower standard deviation is better)"
echo "4. Resource efficiency and stability"
echo
echo -e "${YELLOW}EXPECTED ADVANTAGES OF GO VERSION:${NC}"
echo "• Better handling of concurrent API calls (goroutines vs event loop)"
echo "• More predictable performance under load"
echo "• Lower memory overhead for many concurrent operations"
echo "• Better throughput for API-heavy endpoints" 
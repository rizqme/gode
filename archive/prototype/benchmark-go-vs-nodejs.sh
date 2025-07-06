#!/bin/bash

# 🚨 Go-Powered External API vs Node.js Event Loop Benchmark
# This compares Go goroutines handling external APIs vs Node.js setTimeout promises

echo "🚨 GO-POWERED EXTERNAL API VS NODE.JS BENCHMARK"
echo "================================================="
echo "🔥 Go Version: External API calls handled by Go goroutines"
echo "🔥 Node.js Version: External API calls handled by setTimeout promises"
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

# Test specific endpoint performance
test_endpoint_performance() {
    local name=$1
    local endpoint=$2
    local description=$3
    
    echo -e "${BLUE}Testing ${name} - ${description}${NC}"
    echo "Endpoint: ${endpoint}"
    
    # Test with moderate load
    echo "📊 Performance Test (4 threads, 10 connections, 10 seconds):"
    wrk -t4 -c10 -d10s --latency http://localhost:8080${endpoint} | grep -E "(Requests/sec|Latency|requests in)"
    
    echo
}

# Test concurrent data processing
test_concurrent_processing() {
    local name=$1
    
    echo -e "${BLUE}Testing ${name} - Concurrent Data Processing${NC}"
    
    # Test different batch sizes
    for batch_size in 100 500 1000; do
        echo "📊 Processing ${batch_size} records:"
        
        start_time=$(date +%s%N)
        response=$(curl -s -X POST http://localhost:8080/api/process-data \
            -H "Content-Type: application/json" \
            -d "{\"batch_size\": ${batch_size}}" \
            --max-time 30)
        end_time=$(date +%s%N)
        
        duration=$(( (end_time - start_time) / 1000000 ))  # Convert to ms
        
        echo "  Total time: ${duration}ms"
        echo "  Response: ${response}" | head -c 100
        echo "..."
        echo
    done
}

# Test mixed workload performance
test_mixed_workload() {
    local name=$1
    
    echo -e "${BLUE}Testing ${name} - Mixed Workload (API + CPU)${NC}"
    
    for iterations in 10 50 100; do
        echo "📊 Testing ${iterations} iterations:"
        
        start_time=$(date +%s%N)
        response=$(curl -s -X POST http://localhost:8080/api/heavy-computation \
            -H "Content-Type: application/json" \
            -d "{\"iterations\": ${iterations}}" \
            --max-time 60)
        end_time=$(date +%s%N)
        
        duration=$(( (end_time - start_time) / 1000000 ))  # Convert to ms
        
        echo "  Total time: ${duration}ms"
        echo "  Response: ${response}" | head -c 100
        echo "..."
        echo
    done
}

# Test health check responsiveness during load
test_health_under_load() {
    local name=$1
    
    echo -e "${BLUE}Testing ${name} - Health Check Under Load${NC}"
    
    # Start heavy load in background
    echo "🔥 Starting heavy load (data processing)..."
    curl -s -X POST http://localhost:8080/api/process-data \
        -H "Content-Type: application/json" \
        -d "{\"batch_size\": 1000}" \
        --max-time 30 > /dev/null 2>&1 &
    
    # Start another heavy load
    curl -s -X POST http://localhost:8080/api/process-data \
        -H "Content-Type: application/json" \
        -d "{\"batch_size\": 1000}" \
        --max-time 30 > /dev/null 2>&1 &
    
    sleep 2  # Let load build up
    
    # Test health check during load
    echo "🏥 Health check response times during load:"
    for i in {1..10}; do
        start_time=$(date +%s%N)
        curl -s http://localhost:8080/health > /dev/null
        end_time=$(date +%s%N)
        duration=$(( (end_time - start_time) / 1000000 ))  # Convert to ms
        echo "  Health check #${i}: ${duration}ms"
        sleep 1
    done
    
    echo
}

echo "🚀 TESTING GODE (Go-Powered External APIs)"
echo "==========================================="

# Start Gode server with Go-powered external APIs
echo "Starting Gode server with Go-powered external APIs..."
timeout 180 go run main.go go-api-overload.js > /dev/null 2>&1 &
GODE_PID=$!

if wait_for_server; then
    test_endpoint_performance "GODE" "/api/user-profile/test123" "15 Go goroutine API calls"
    test_endpoint_performance "GODE" "/api/realtime-data" "250 concurrent Go API calls"
    test_concurrent_processing "GODE"
    test_mixed_workload "GODE"
    test_health_under_load "GODE"
    
    echo -e "${GREEN}✅ Gode (Go-powered) testing completed${NC}"
else
    echo -e "${RED}❌ Gode server failed to start${NC}"
fi

# Kill Gode server
kill $GODE_PID 2>/dev/null || true
sleep 3

echo
echo "🚀 TESTING NODE.JS (setTimeout-based External APIs)"
echo "===================================================="

# Start Node.js server with setTimeout-based external APIs
echo "Starting Node.js server with setTimeout-based external APIs..."
cd baseline
timeout 180 node api-overload.js > /dev/null 2>&1 &
NODE_PID=$!
cd ..

if wait_for_server; then
    test_endpoint_performance "NODE.JS" "/api/user-profile/test123" "15 setTimeout promise API calls"
    test_endpoint_performance "NODE.JS" "/api/realtime-data" "250 concurrent setTimeout promises"
    test_concurrent_processing "NODE.JS"
    test_mixed_workload "NODE.JS"
    test_health_under_load "NODE.JS"
    
    echo -e "${GREEN}✅ Node.js (setTimeout-based) testing completed${NC}"
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
echo -e "${YELLOW}EXPECTED ADVANTAGES OF GO VERSION:${NC}"
echo "1. Better concurrency handling (goroutines vs event loop)"
echo "2. Lower memory overhead for many concurrent operations"
echo "3. Better health check responsiveness under load"
echo "4. More predictable performance under stress"
echo
echo -e "${YELLOW}WHAT TO OBSERVE:${NC}"
echo "• Response times for endpoints with many external API calls"
echo "• Health check latency during heavy processing"
echo "• Memory usage and stability under concurrent load"
echo "• Overall system responsiveness" 
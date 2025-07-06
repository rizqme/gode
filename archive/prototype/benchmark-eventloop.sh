#!/bin/bash

# 🚨 Event Loop Overload Benchmark 
# This demonstrates how Node.js event loop gets saturated with pending promises

echo "🚨 EVENT LOOP OVERLOAD BENCHMARK"
echo "=================================="
echo

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
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

# Test heavy endpoint under load
test_heavy_endpoint() {
    local name=$1
    local endpoint=$2
    
    echo -e "${YELLOW}Testing ${name} - Heavy Load${NC}"
    echo "Endpoint: ${endpoint}"
    
    # First test with low concurrency
    echo "📊 Light Load (2 threads, 5 connections):"
    wrk -t2 -c5 -d10s --latency http://localhost:8080${endpoint} | head -15
    
    echo
    echo "📊 Medium Load (4 threads, 20 connections):"
    wrk -t4 -c20 -d10s --latency http://localhost:8080${endpoint} | head -15
    
    echo
    echo "📊 Heavy Load (8 threads, 50 connections):"
    wrk -t8 -c50 -d10s --latency http://localhost:8080${endpoint} | head -15
    
    echo
}

# Test health check responsiveness during load
test_health_during_load() {
    local name=$1
    local heavy_endpoint=$2
    
    echo -e "${YELLOW}Testing ${name} - Health Check During Load${NC}"
    
    # Start heavy load in background
    echo "🔥 Starting heavy load on ${heavy_endpoint}..."
    wrk -t8 -c50 -d30s http://localhost:8080${heavy_endpoint} > /dev/null 2>&1 &
    local heavy_pid=$!
    
    sleep 3  # Let load build up
    
    # Test health check during load
    echo "🏥 Testing health check responsiveness during load:"
    for i in {1..10}; do
        start_time=$(date +%s%N)
        curl -s http://localhost:8080/health > /dev/null
        end_time=$(date +%s%N)
        duration=$(( (end_time - start_time) / 1000000 ))  # Convert to ms
        echo "  Health check #${i}: ${duration}ms"
        sleep 1
    done
    
    # Kill the heavy load
    kill $heavy_pid 2>/dev/null || true
    sleep 2
    
    echo
}

# Test data processing endpoint (creates thousands of promises)
test_data_processing() {
    local name=$1
    
    echo -e "${YELLOW}Testing ${name} - Data Processing (Promise Hell)${NC}"
    
    # Test with different batch sizes
    for batch_size in 100 500 1000; do
        echo "📊 Processing ${batch_size} records:"
        start_time=$(date +%s%N)
        
        response=$(curl -s -X POST http://localhost:8080/api/process-data \
            -H "Content-Type: application/json" \
            -d "{\"batch_size\": ${batch_size}, \"start_time\": $(date +%s%N)}")
        
        end_time=$(date +%s%N)
        duration=$(( (end_time - start_time) / 1000000 ))  # Convert to ms
        
        echo "  Response time: ${duration}ms"
        echo "  Response: ${response}"
        echo
    done
}

# Test concurrent data processing (the real killer)
test_concurrent_processing() {
    local name=$1
    
    echo -e "${YELLOW}Testing ${name} - Concurrent Data Processing${NC}"
    
    # Launch multiple processing requests simultaneously
    echo "🔥 Launching 5 concurrent data processing requests..."
    
    for i in {1..5}; do
        {
            start_time=$(date +%s%N)
            response=$(curl -s -X POST http://localhost:8080/api/process-data \
                -H "Content-Type: application/json" \
                -d "{\"batch_size\": 500, \"start_time\": $(date +%s%N)}")
            end_time=$(date +%s%N)
            duration=$(( (end_time - start_time) / 1000000 ))
            echo "  Request #${i}: ${duration}ms - ${response}"
        } &
    done
    
    # Wait for all requests to complete
    wait
    echo
}

echo "🚀 TESTING GODE (Goja + Gin)"
echo "=============================="

# Start Gode server
echo "Starting Gode server..."
timeout 120 go run main.go api-overload.js > /dev/null 2>&1 &
GODE_PID=$!

if wait_for_server; then
    test_heavy_endpoint "GODE" "/api/user-profile/test123"
    test_health_during_load "GODE" "/api/user-profile/test123"
    test_data_processing "GODE"
    test_concurrent_processing "GODE"
    
    echo -e "${GREEN}✅ Gode testing completed${NC}"
else
    echo -e "${RED}❌ Gode server failed to start${NC}"
fi

# Kill Gode server
kill $GODE_PID 2>/dev/null || true
sleep 3

echo
echo "🚀 TESTING NODE.JS (Express)"
echo "============================="

# Start Node.js server
echo "Starting Node.js server..."
cd baseline
timeout 120 node api-overload.js > /dev/null 2>&1 &
NODE_PID=$!
cd ..

if wait_for_server; then
    test_heavy_endpoint "NODE.JS" "/api/user-profile/test123"
    test_health_during_load "NODE.JS" "/api/user-profile/test123"
    test_data_processing "NODE.JS"
    test_concurrent_processing "NODE.JS"
    
    echo -e "${GREEN}✅ Node.js testing completed${NC}"
else
    echo -e "${RED}❌ Node.js server failed to start${NC}"
fi

# Kill Node.js server
kill $NODE_PID 2>/dev/null || true

echo
echo "🏁 BENCHMARK COMPLETE"
echo "===================="
echo "Key observations to look for:"
echo "1. Health check response times under load"
echo "2. Data processing performance with many promises"
echo "3. Concurrent request handling capability"
echo "4. Overall system stability under stress" 
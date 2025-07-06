#!/bin/bash

# 🧪 Gode vs Node.js Benchmark Script
# This script benchmarks both Gode (Goja + Gin) and Node.js (Express) servers

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
PORT=8080
DURATION=10s
THREADS=20
CONNECTIONS=100
WARMUP_TIME=3

echo -e "${BLUE}🧪 Gode vs Node.js Benchmark${NC}"
echo "=============================================="
echo "Configuration:"
echo "  - Port: $PORT"
echo "  - Duration: $DURATION"
echo "  - Threads: $THREADS"
echo "  - Connections: $CONNECTIONS"
echo "  - Warmup: ${WARMUP_TIME}s"
echo ""

# Function to check if a command exists
check_command() {
    if ! command -v $1 &> /dev/null; then
        echo -e "${RED}Error: $1 is not installed${NC}"
        exit 1
    fi
}

# Function to wait for server to be ready
wait_for_server() {
    local server_name=$1
    local max_attempts=30
    local attempt=0
    
    echo -e "${YELLOW}Waiting for $server_name to be ready...${NC}"
    
    while ! curl -s http://localhost:$PORT/health > /dev/null 2>&1; do
        attempt=$((attempt + 1))
        if [ $attempt -ge $max_attempts ]; then
            echo -e "${RED}Error: $server_name failed to start after $max_attempts attempts${NC}"
            exit 1
        fi
        sleep 1
    done
    
    echo -e "${GREEN}$server_name is ready!${NC}"
    sleep $WARMUP_TIME
}

# Function to kill background processes
cleanup() {
    echo -e "${YELLOW}Cleaning up...${NC}"
    pkill -f "go run main.go" || true
    pkill -f "node baseline/app.js" || true
    sleep 2
}

# Function to run benchmark
run_benchmark() {
    local server_name=$1
    local endpoint=$2
    local test_name=$3
    
    echo -e "${BLUE}Running $test_name benchmark on $server_name...${NC}"
    
    # Run wrk benchmark
    local result=$(wrk -t$THREADS -c$CONNECTIONS -d$DURATION --latency http://localhost:$PORT$endpoint 2>&1)
    
    # Extract key metrics
    local rps=$(echo "$result" | grep "Requests/sec" | awk '{print $2}')
    local latency_avg=$(echo "$result" | grep "Latency" | awk '{print $2}')
    local latency_max=$(echo "$result" | grep "Latency" | awk '{print $4}')
    local total_requests=$(echo "$result" | grep "requests in" | awk '{print $1}')
    
    echo "  Requests/sec: $rps"
    echo "  Avg Latency: $latency_avg"
    echo "  Max Latency: $latency_max"
    echo "  Total Requests: $total_requests"
    echo ""
    
    # Store results for comparison
    echo "$server_name,$endpoint,$test_name,$rps,$latency_avg,$latency_max,$total_requests" >> benchmark_results.csv
}

# Function to test streaming
test_streaming() {
    local server_name=$1
    
    echo -e "${BLUE}Testing streaming on $server_name...${NC}"
    
    # Test streaming endpoint
    local start_time=$(date +%s.%N)
    local response=$(curl -s -w "%{time_total}" http://localhost:$PORT/stream)
    local end_time=$(date +%s.%N)
    
    local total_time=$(echo "$response" | tail -n1)
    local chunk_count=$(echo "$response" | head -n -1 | wc -l)
    
    echo "  Total Time: ${total_time}s"
    echo "  Chunks Received: $chunk_count"
    echo "  Streaming: $([ $chunk_count -eq 5 ] && echo "✅ OK" || echo "❌ FAILED")"
    echo ""
}

# Function to measure memory usage
measure_memory() {
    local server_name=$1
    local pid=$2
    
    echo -e "${BLUE}Measuring memory usage for $server_name...${NC}"
    
    # Wait a bit for the server to stabilize
    sleep 2
    
    # Get memory usage (RSS in KB)
    local memory_kb=$(ps -o rss= -p $pid 2>/dev/null || echo "0")
    local memory_mb=$((memory_kb / 1024))
    
    echo "  Memory Usage: ${memory_mb}MB (${memory_kb}KB RSS)"
    echo ""
}

# Check required tools
echo -e "${YELLOW}Checking required tools...${NC}"
check_command "go"
check_command "node"
check_command "npm"
check_command "wrk"
check_command "curl"

# Setup trap for cleanup
trap cleanup EXIT

# Install Node.js dependencies
echo -e "${YELLOW}Installing Node.js dependencies...${NC}"
cd baseline && npm install --silent && cd ..

# Initialize results file
echo "Server,Endpoint,Test,RPS,AvgLatency,MaxLatency,TotalRequests" > benchmark_results.csv

# Benchmark Gode (Goja + Gin)
echo -e "${GREEN}========== Benchmarking Gode (Goja + Gin) ==========${NC}"

# Start Gode server
go run main.go example.js &
GODE_PID=$!
wait_for_server "Gode"

# Measure initial memory
measure_memory "Gode" $GODE_PID

# Run benchmarks
run_benchmark "Gode" "/ping" "JSON"
run_benchmark "Gode" "/health" "Health"
test_streaming "Gode"

# Stop Gode server
kill $GODE_PID || true
sleep 2

# Benchmark Node.js (Express)
echo -e "${GREEN}========== Benchmarking Node.js (Express) ==========${NC}"

# Start Node.js server
cd baseline && node app.js &
NODEJS_PID=$!
cd ..
wait_for_server "Node.js"

# Measure initial memory
measure_memory "Node.js" $NODEJS_PID

# Run benchmarks
run_benchmark "Node.js" "/ping" "JSON"
run_benchmark "Node.js" "/health" "Health"
test_streaming "Node.js"

# Stop Node.js server
kill $NODEJS_PID || true
sleep 2

# Display results summary
echo -e "${GREEN}========== Benchmark Results Summary ==========${NC}"
echo ""
echo "📊 Detailed results saved to: benchmark_results.csv"
echo ""

# Show CSV results in a nice format
if command -v column &> /dev/null; then
    echo "Results Table:"
    column -t -s, benchmark_results.csv
else
    echo "Results (CSV format):"
    cat benchmark_results.csv
fi

echo ""
echo -e "${GREEN}✅ Benchmark completed successfully!${NC}"
echo ""
echo "🚀 To run individual tests:"
echo "  Gode:    go run main.go example.js"
echo "  Node.js: cd baseline && node app.js"
echo ""
echo "📈 To run performance tests:"
echo "  wrk -t20 -c100 -d10s http://localhost:8080/ping"
echo "  curl -N http://localhost:8080/stream" 
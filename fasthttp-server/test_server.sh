#!/bin/bash

# FastHTTP Server Test Script

echo "Testing FastHTTP Server..."

# Colors
GREEN='\033[0;32m'
RED='\033[0;31m'
NC='\033[0m'

print_test() { echo -e "${GREEN}[TEST]${NC} $1"; }
print_fail() { echo -e "${RED}[FAIL]${NC} $1"; }

# Check if server is built
if [ ! -f "./fasthttp-server" ]; then
    echo "Building server..."
    go build -o fasthttp-server .
fi

# Start server in background
echo "Starting server..."
./fasthttp-server server -p 8080 -l debug &
SERVER_PID=$!
echo "Server PID: $SERVER_PID"

# Wait for server to start
sleep 3

# Test endpoints
print_test "Testing endpoints..."

echo "=== Root endpoint ==="
curl -i http://localhost:8080/

echo -e "\n=== Health endpoint ==="
curl -i http://localhost:8080/health

echo -e "\n=== Status endpoint ==="
curl -i http://localhost:8080/api/v1/status

echo -e "\n=== 404 test ==="
curl -i http://localhost:8080/notfound

# Test request ID tracking
echo -e "\n=== Request ID tracking ==="
for i in {1..3}; do
    echo "Request $i:"
    curl -s http://localhost:8080/health | grep -o '"request_id":"[^"]*"'
done

# Test concurrent requests
echo -e "\n=== Concurrent requests ==="
for i in {1..5}; do
    curl -s http://localhost:8080/health &
done
wait

# Check logs
echo -e "\n=== Log files ==="
ls -la logs/
echo -e "\nRecent log entries:"
tail -10 logs/server_*.log

# Stop server
echo -e "\nStopping server..."
kill $SERVER_PID
wait $SERVER_PID 2>/dev/null

# Check final logs
echo -e "\nFinal log entries:"
tail -5 logs/server_*.log

echo -e "\n${GREEN}Test completed!${NC}"
echo "Check logs directory: ls logs/"

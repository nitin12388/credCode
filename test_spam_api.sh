#!/bin/bash

echo "=== Testing Spam Detection API ==="
echo ""

# Start server in background
echo "Starting server..."
cd /Users/admin_1/credCode
go run cmd/server/main.go &
SERVER_PID=$!

# Wait for server to start
sleep 2

# Test 1: Health check
echo "Test 1: Health check"
curl -s http://localhost:8080/health | jq .
echo ""

# Test 2: Get registered rules
echo "Test 2: Get registered rules"
curl -s http://localhost:8080/api/v1/spam/rules | jq .
echo ""

# Test 3: Detect spam for a phone number
echo "Test 3: Detect spam for phone 7379037972"
curl -s -X POST http://localhost:8080/api/v1/spam/detect \
  -H "Content-Type: application/json" \
  -d '{"phone_number": "7379037972"}' | jq .
echo ""

# Test 4: Detect spam for unknown number
echo "Test 4: Detect spam for unknown number 9999999999"
curl -s -X POST http://localhost:8080/api/v1/spam/detect \
  -H "Content-Type: application/json" \
  -d '{"phone_number": "9999999999"}' | jq .
echo ""

# Kill server
echo "Stopping server..."
kill $SERVER_PID
wait $SERVER_PID 2>/dev/null

echo "=== Tests Complete ==="


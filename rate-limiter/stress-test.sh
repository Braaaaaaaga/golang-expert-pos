#!/bin/bash

BASE_URL="http://localhost:8080"
ENDPOINT="/test"

echo "🔥 Rate Limiter Stress Test 🔥"
echo "================================"
echo ""
echo "Test 1: IP Rate Limiting (10 req/s limit)"
echo "Sending 15 requests from same IP..."
for i in {1..15}; do
    response=$(curl -s -w "\n%{http_code}" $BASE_URL$ENDPOINT)
    status_code=$(echo "$response" | tail -n1)
    echo "Request $i: HTTP $status_code"
    
    if [ $i -eq 10 ]; then
        echo "  ⏰ Waiting 1 second..."
        sleep 1
    fi
done

echo ""
echo "Test 2: Token-based Rate Limiting (100 req/s limit)"
echo "Sending 105 requests with valid token..."
for i in {1..105}; do
    response=$(curl -s -w "\n%{http_code}" -H "API_KEY: abc123" $BASE_URL$ENDPOINT)
    status_code=$(echo "$response" | tail -n1)
    
    if [ $((i % 20)) -eq 0 ]; then
        echo "Request $i: HTTP $status_code"
    fi
    
    if [ $i -eq 100 ]; then
        echo "  ⏰ Waiting 1 second..."
        sleep 1
    fi
done

echo ""
echo "Test 3: Invalid Token (falls back to IP limit)"
echo "Sending 12 requests with invalid token..."
for i in {1..12}; do
    response=$(curl -s -w "\n%{http_code}" -H "API_KEY: invalid-token" $BASE_URL$ENDPOINT)
    status_code=$(echo "$response" | tail -n1)
    echo "Request $i: HTTP $status_code"
    
    if [ $i -eq 10 ]; then
        echo "  ⏰ Waiting 1 second..."
        sleep 1
    fi
done

echo ""
echo "Test 4: Concurrent requests from different IPs"
echo "Sending 5 concurrent requests from different IPs..."

# Function to make request with custom IP header
make_request() {
    local ip=$1
    local req_num=$2
    response=$(curl -s -w "\n%{http_code}" -H "X-Forwarded-For: $ip" $BASE_URL$ENDPOINT)
    status_code=$(echo "$response" | tail -n1)
    echo "Request from IP $ip ($req_num): HTTP $status_code"
}

# Launch concurrent requests
for i in {1..5}; do
    make_request "192.168.1.$i" $i &
done

# Wait for all background jobs to complete
wait

echo ""
echo "Test 5: Token priority over IP"
echo "First exhaust IP limit, then use token..."

# Exhaust IP limit
echo "Exhausting IP limit..."
for i in {1..11}; do
    response=$(curl -s -w "\n%{http_code}" $BASE_URL$ENDPOINT)
    status_code=$(echo "$response" | tail -n1)
    if [ $i -le 3 ] || [ $i -gt 10 ]; then
        echo "Request $i (no token): HTTP $status_code"
    fi
done

# Now use token (should work despite IP being blocked)
echo "Using valid token after IP exhaustion..."
response=$(curl -s -w "\n%{http_code}" -H "API_KEY: abc123" $BASE_URL$ENDPOINT)
status_code=$(echo "$response" | tail -n1)
echo "Request with token: HTTP $status_code"

echo ""
echo "✅ Stress test completed!"
echo ""
echo "Expected results:"
echo "- Test 1: First 10 requests should return 200, rest should return 429"
echo "- Test 2: First 100 requests should return 200, rest should return 429"
echo "- Test 3: First 10 requests should return 200, rest should return 429"
echo "- Test 4: All requests should return 200 (different IPs)"
echo "- Test 5: Token request should return 200 despite IP being blocked"
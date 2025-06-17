#!/bin/bash
set -e

echo "🔐 Testing Authentication Endpoints..."
echo ""

# Test 1: User Registration
echo "1. Testing user registration..."
TIMESTAMP1=$(date +%s)
TEST_EMAIL1="testuser${TIMESTAMP1}@example.com"
echo "Using test email: $TEST_EMAIL1"

REGISTER_RESPONSE=$(curl -s -X POST http://localhost:8080/api/v1/auth/register \
    -H "Content-Type: application/json" \
    -d "{\"email\":\"$TEST_EMAIL1\",\"password\":\"testpass123\",\"first_name\":\"Test\",\"last_name\":\"User\",\"phone\":\"+1234567890\"}")

echo "Registration Response: $REGISTER_RESPONSE"
if echo "$REGISTER_RESPONSE" | grep -q "access_token"; then
    echo "✅ User registration successful"
else
    echo "❌ User registration failed"
    exit 1
fi
echo ""

# Test 2: User Login
echo "2. Testing user login..."
sleep 1
TIMESTAMP2=$(date +%s)
TEST_EMAIL2="logintest${TIMESTAMP2}@example.com"

# First register a user for login test
curl -s -X POST http://localhost:8080/api/v1/auth/register \
    -H "Content-Type: application/json" \
    -d "{\"email\":\"$TEST_EMAIL2\",\"password\":\"testpass123\",\"first_name\":\"Test\",\"last_name\":\"User\",\"phone\":\"+1234567890\"}" > /dev/null

LOGIN_RESPONSE=$(curl -s -X POST http://localhost:8080/api/v1/auth/login \
    -H "Content-Type: application/json" \
    -d "{\"email\":\"$TEST_EMAIL2\",\"password\":\"testpass123\"}")

echo "Login Response: $LOGIN_RESPONSE"
if echo "$LOGIN_RESPONSE" | grep -q "access_token"; then
    echo "✅ User login successful"
else
    echo "❌ User login failed"
    exit 1
fi
echo ""

# Test 3: Invalid Login
echo "3. Testing invalid login..."
sleep 1
TIMESTAMP3=$(date +%s)
TEST_EMAIL3="invalidtest${TIMESTAMP3}@example.com"

INVALID_LOGIN_RESPONSE=$(curl -s -X POST http://localhost:8080/api/v1/auth/login \
    -H "Content-Type: application/json" \
    -d "{\"email\":\"$TEST_EMAIL3\",\"password\":\"wrongpassword\"}")

echo "Invalid Login Response: $INVALID_LOGIN_RESPONSE"
if echo "$INVALID_LOGIN_RESPONSE" | grep -q "error"; then
    echo "✅ Invalid login properly rejected"
else
    echo "❌ Invalid login not properly handled"
    exit 1
fi
echo ""

echo "🔐 Authentication tests completed successfully!"

#!/bin/bash
set -e

echo "🚫 Testing Error Scenarios and Security..."
echo ""

# Test 1: Unauthorized Access
echo "1. Testing unauthorized access to protected endpoint..."
UNAUTH_RESPONSE=$(curl -s -X GET http://localhost:8080/api/v1/profile)
echo "Unauthorized Response: $UNAUTH_RESPONSE"
if echo "$UNAUTH_RESPONSE" | grep -q "error"; then
    echo "✅ Unauthorized access properly blocked"
else
    echo "❌ Unauthorized access not properly blocked"
    exit 1
fi
echo ""

# Test 2: Invalid JWT Token
echo "2. Testing invalid JWT token..."
INVALID_TOKEN_RESPONSE=$(curl -s -X GET http://localhost:8080/api/v1/profile \
    -H "Authorization: Bearer invalid-token-here")
echo "Invalid Token Response: $INVALID_TOKEN_RESPONSE"
if echo "$INVALID_TOKEN_RESPONSE" | grep -q "error"; then
    echo "✅ Invalid token properly rejected"
else
    echo "❌ Invalid token not properly handled"
    exit 1
fi
echo ""

# Test 3: Duplicate User Registration
echo "3. Testing duplicate user registration..."
TIMESTAMP=$(date +%s)
DUPLICATE_EMAIL="duplicate${TIMESTAMP}@example.com"

# Register first user
curl -s -X POST http://localhost:8080/api/v1/auth/register \
    -H "Content-Type: application/json" \
    -d "{\"email\":\"$DUPLICATE_EMAIL\",\"password\":\"pass12345678\",\"first_name\":\"First\",\"last_name\":\"User\"}" > /dev/null

# Try to register same email again
DUPLICATE_RESPONSE=$(curl -s -X POST http://localhost:8080/api/v1/auth/register \
    -H "Content-Type: application/json" \
    -d "{\"email\":\"$DUPLICATE_EMAIL\",\"password\":\"pass12345678\",\"first_name\":\"Second\",\"last_name\":\"User\"}")

echo "Duplicate Registration Response: $DUPLICATE_RESPONSE"
if echo "$DUPLICATE_RESPONSE" | grep -q "error"; then
    echo "✅ Duplicate registration properly prevented"
else
    echo "❌ Duplicate registration not properly handled"
    exit 1
fi
echo ""

# Test 4: Weak Password Rejection
echo "4. Testing weak password rejection..."
sleep 1
TIMESTAMP2=$(date +%s)
WEAK_EMAIL="weak${TIMESTAMP2}@example.com"

WEAK_PASSWORD_RESPONSE=$(curl -s -X POST http://localhost:8080/api/v1/auth/register \
    -H "Content-Type: application/json" \
    -d "{\"email\":\"$WEAK_EMAIL\",\"password\":\"123\",\"first_name\":\"Weak\",\"last_name\":\"Password\"}")

echo "Weak Password Response: $WEAK_PASSWORD_RESPONSE"
if echo "$WEAK_PASSWORD_RESPONSE" | grep -q "error"; then
    echo "✅ Weak password properly rejected"
else
    echo "❌ Weak password not properly handled"
    exit 1
fi
echo ""

# Test 5: Access to Non-existent User
echo "5. Testing access to non-existent user..."
sleep 1
TIMESTAMP3=$(date +%s)
CHECKER_EMAIL="checker${TIMESTAMP3}@example.com"

REGISTER_RESPONSE=$(curl -s -X POST http://localhost:8080/api/v1/auth/register \
    -H "Content-Type: application/json" \
    -d "{\"email\":\"$CHECKER_EMAIL\",\"password\":\"checkpass123\",\"first_name\":\"Check\",\"last_name\":\"User\"}")

JWT_TOKEN=$(echo "$REGISTER_RESPONSE" | jq -r '.access_token')

NONEXISTENT_RESPONSE=$(curl -s -X GET http://localhost:8080/api/v1/users/nonexistent-id \
    -H "Authorization: Bearer $JWT_TOKEN")

echo "Non-existent User Response: $NONEXISTENT_RESPONSE"
if echo "$NONEXISTENT_RESPONSE" | grep -q "error"; then
    echo "✅ Non-existent user properly handled"
else
    echo "❌ Non-existent user not properly handled"
    exit 1
fi
echo ""

echo "🚫 Error handling tests completed successfully!"

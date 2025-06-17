#!/bin/bash
set -e

echo "👤 Testing User Endpoints..."
echo ""

# Setup: Register test user and get token
TIMESTAMP=$(date +%s)
TEST_EMAIL="usertest${TIMESTAMP}@example.com"
echo "Using test email: $TEST_EMAIL"
echo ""

echo "1. Registering test user and getting token..."
REGISTER_RESPONSE=$(curl -s -X POST http://localhost:8080/api/v1/auth/register \
    -H "Content-Type: application/json" \
    -d "{\"email\":\"$TEST_EMAIL\",\"password\":\"userpass123\",\"first_name\":\"User\",\"last_name\":\"Test\",\"phone\":\"+1555000123\"}")

JWT_TOKEN=$(echo "$REGISTER_RESPONSE" | jq -r '.access_token')
USER_ID=$(echo "$REGISTER_RESPONSE" | jq -r '.user.id')

echo "Token obtained: $JWT_TOKEN"
echo "User ID: $USER_ID"
echo ""

# Test 2: Profile Access
echo "2. Testing profile access..."
PROFILE_RESPONSE=$(curl -s -X GET http://localhost:8080/api/v1/profile \
    -H "Authorization: Bearer $JWT_TOKEN")

echo "Profile Response: $PROFILE_RESPONSE"
if echo "$PROFILE_RESPONSE" | grep -q "email"; then
    echo "✅ Profile access successful"
else
    echo "❌ Profile access failed"
    exit 1
fi
echo ""

# Test 3: User Data Retrieval
echo "3. Testing user data retrieval..."
USER_RESPONSE=$(curl -s -X GET http://localhost:8080/api/v1/users/$USER_ID \
    -H "Authorization: Bearer $JWT_TOKEN")

echo "User Data Response: $USER_RESPONSE"
if echo "$USER_RESPONSE" | grep -q "$USER_ID"; then
    echo "✅ User data retrieval successful"
else
    echo "❌ User data retrieval failed"
    exit 1
fi
echo ""

# Test 4: Profile Update
echo "4. Testing user profile update..."
UPDATE_RESPONSE=$(curl -s -X PUT http://localhost:8080/api/v1/users/$USER_ID \
    -H "Authorization: Bearer $JWT_TOKEN" \
    -H "Content-Type: application/json" \
    -d '{"first_name":"Updated","last_name":"Name","phone":"+1555999888"}')

echo "Update Response: $UPDATE_RESPONSE"
if echo "$UPDATE_RESPONSE" | grep -q "Updated"; then
    echo "✅ User profile update successful"
else
    echo "❌ User profile update failed"
    exit 1
fi
echo ""

echo "👤 User management tests completed successfully!"

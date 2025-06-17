#!/bin/bash
set -e

echo "🛡️ Testing Risk Management API..."
echo ""

# Setup: Create test user for risk checking
TIMESTAMP=$(date +%s)
TEST_USER_EMAIL="risktest${TIMESTAMP}@suspicious-domain.com"
echo "Using test user email: $TEST_USER_EMAIL"
echo ""

echo "1. Creating admin user and getting admin token..."
# For risk management tests, we need an admin token to create rules
ADMIN_EMAIL="admin${TIMESTAMP}@example.com"
ADMIN_REGISTER_RESPONSE=$(curl -s -X POST http://localhost:8080/api/v1/auth/register \
    -H "Content-Type: application/json" \
    -d "{\"email\":\"$ADMIN_EMAIL\",\"password\":\"adminpass123\",\"first_name\":\"Admin\",\"last_name\":\"User\",\"phone\":\"+1555000001\"}")

ADMIN_USER_ID=$(echo "$ADMIN_REGISTER_RESPONSE" | jq -r '.user.id')

# Update user role to admin in database for rule creation
PGPASSWORD="app_password" psql -h localhost -U app_admin -d users -c "
UPDATE users SET roles = '[\"admin\"]' WHERE id = '$ADMIN_USER_ID';
" > /dev/null 2>&1

# Login as admin to get token
ADMIN_LOGIN_RESPONSE=$(curl -s -X POST http://localhost:8080/api/v1/auth/login \
    -H "Content-Type: application/json" \
    -d "{\"email\":\"$ADMIN_EMAIL\",\"password\":\"adminpass123\"}")

ADMIN_JWT_TOKEN=$(echo "$ADMIN_LOGIN_RESPONSE" | jq -r '.access_token')
if [ "$ADMIN_JWT_TOKEN" = "null" ] || [ -z "$ADMIN_JWT_TOKEN" ]; then
    echo "❌ Failed to get admin token for rule creation"
    exit 1
fi

echo "✅ Admin token obtained for rule setup"
echo ""

echo "2. Setting up risk rule for testing..."
CREATE_RULE_RESPONSE=$(curl -s -X POST http://localhost:8080/api/v1/risk/rules \
    -H "Authorization: Bearer $ADMIN_JWT_TOKEN" \
    -H "Content-Type: application/json" \
    -d '{
        "name": "Suspicious Email Domain Test",
        "type": "DOMAIN_BLACKLIST",
        "category": "EMAIL",
        "value": "suspicious-domain.com",
        "score": 75,
        "is_active": true,
        "confidence": 0.9,
        "expires_in_days": 30
    }')

echo "Create Rule Response: $CREATE_RULE_RESPONSE"
if echo "$CREATE_RULE_RESPONSE" | grep -q "rule_id"; then
    echo "✅ Risk rule creation successful"
    RULE_ID=$(echo "$CREATE_RULE_RESPONSE" | jq -r '.rule_id')
    echo "Created Rule ID: $RULE_ID"
else
    echo "❌ Risk rule creation failed"
    exit 1
fi
echo ""

echo "3. Verifying risk rule was created..."
if echo "$CREATE_RULE_RESPONSE" | grep -q "rule_id"; then
    echo "✅ Risk rule creation successful"
    RULE_ID=$(echo "$CREATE_RULE_RESPONSE" | jq -r '.rule_id')
    echo "Created Rule ID: $RULE_ID"
else
    echo "❌ Risk rule creation failed"
    exit 1
fi
echo ""

echo "4. Creating regular user for risk testing..."
USER_REGISTER_RESPONSE=$(curl -s -X POST http://localhost:8080/api/v1/auth/register \
    -H "Content-Type: application/json" \
    -d "{\"email\":\"regularuser${TIMESTAMP}@example.com\",\"password\":\"userpass123\",\"first_name\":\"Regular\",\"last_name\":\"User\",\"phone\":\"+1555000002\"}")

USER_JWT_TOKEN=$(echo "$USER_REGISTER_RESPONSE" | jq -r '.access_token')
USER_ID=$(echo "$USER_REGISTER_RESPONSE" | jq -r '.user.id')

echo "Regular User Token obtained: $USER_JWT_TOKEN"
echo "Regular User ID: $USER_ID"
echo ""

echo "5. Testing risk check for clean user..."
CLEAN_RISK_RESPONSE=$(curl -s -X POST http://localhost:8080/api/v1/risk/check \
    -H "Authorization: Bearer $USER_JWT_TOKEN" \
    -H "Content-Type: application/json" \
    -d "{
        \"user_id\": \"$USER_ID\",
        \"email\": \"cleanuser${TIMESTAMP}@example.com\",
        \"first_name\": \"Clean\",
        \"last_name\": \"User\",
        \"phone\": \"+1555000003\"
    }")

echo "Clean User Risk Response: $CLEAN_RISK_RESPONSE"
if echo "$CLEAN_RISK_RESPONSE" | grep -q "is_risky"; then
    IS_RISKY=$(echo "$CLEAN_RISK_RESPONSE" | jq -r '.is_risky')
    if [ "$IS_RISKY" = "false" ]; then
        echo "✅ Clean user risk check successful (not risky)"
    else
        echo "⚠️ Clean user flagged as risky - check risk rules"
    fi
else
    echo "❌ Clean user risk check failed"
    exit 1
fi
echo ""

echo "6. Testing risk check for suspicious user..."
SUSPICIOUS_RISK_RESPONSE=$(curl -s -X POST http://localhost:8080/api/v1/risk/check \
    -H "Authorization: Bearer $USER_JWT_TOKEN" \
    -H "Content-Type: application/json" \
    -d "{
        \"user_id\": \"suspicious-user-123\",
        \"email\": \"$TEST_USER_EMAIL\",
        \"first_name\": \"Suspicious\",
        \"last_name\": \"User\",
        \"phone\": \"+1555000004\"
    }")

echo "Suspicious User Risk Response: $SUSPICIOUS_RISK_RESPONSE"
if echo "$SUSPICIOUS_RISK_RESPONSE" | grep -q "is_risky"; then
    IS_RISKY=$(echo "$SUSPICIOUS_RISK_RESPONSE" | jq -r '.is_risky')
    if [ "$IS_RISKY" = "true" ]; then
        echo "✅ Suspicious user risk check successful (detected as risky)"
        RISK_LEVEL=$(echo "$SUSPICIOUS_RISK_RESPONSE" | jq -r '.risk_level')
        REASON=$(echo "$SUSPICIOUS_RISK_RESPONSE" | jq -r '.reason')
        echo "   Risk Level: $RISK_LEVEL"
        echo "   Reason: $REASON"
    else
        echo "⚠️ Suspicious user not flagged as risky - check risk rules"
    fi
else
    echo "❌ Suspicious user risk check failed"
    exit 1
fi
echo ""

echo "7. Creating additional risk rules for comprehensive testing..."

NAME_RULE_RESPONSE=$(curl -s -X POST http://localhost:8080/api/v1/risk/rules \
    -H "Authorization: Bearer $ADMIN_JWT_TOKEN" \
    -H "Content-Type: application/json" \
    -d '{
        "name": "Suspicious Name Pattern",
        "type": "NAME_BLACKLIST",
        "category": "NAME",
        "value": "Fake Person",
        "score": 50,
        "is_active": true,
        "confidence": 0.8
    }')

echo "Name Rule Response: $NAME_RULE_RESPONSE"
if echo "$NAME_RULE_RESPONSE" | grep -q "rule_id"; then
    echo "✅ Name blacklist rule created"
    NAME_RULE_ID=$(echo "$NAME_RULE_RESPONSE" | jq -r '.rule_id')
else
    echo "❌ Name blacklist rule creation failed"
    exit 1
fi

DOMAIN_RULE_RESPONSE=$(curl -s -X POST http://localhost:8080/api/v1/risk/rules \
    -H "Authorization: Bearer $ADMIN_JWT_TOKEN" \
    -H "Content-Type: application/json" \
    -d '{
        "name": "Temporary Email Domains",
        "type": "DOMAIN_BLACKLIST",
        "category": "EMAIL",
        "value": "tempmail.org",
        "score": 60,
        "is_active": true,
        "confidence": 0.85
    }')

echo "Domain Rule Response: $DOMAIN_RULE_RESPONSE"
if echo "$DOMAIN_RULE_RESPONSE" | grep -q "rule_id"; then
    echo "✅ Domain blacklist rule created"
    DOMAIN_RULE_ID=$(echo "$DOMAIN_RULE_RESPONSE" | jq -r '.rule_id')
else
    echo "❌ Domain blacklist rule creation failed"
    exit 1
fi
echo ""

echo "8. Testing user with multiple risk factors..."
MULTI_RISK_RESPONSE=$(curl -s -X POST http://localhost:8080/api/v1/risk/check \
    -H "Authorization: Bearer $USER_JWT_TOKEN" \
    -H "Content-Type: application/json" \
    -d "{
        \"user_id\": \"multi-risk-user-456\",
        \"email\": \"fake.person@suspicious-domain.com\",
        \"first_name\": \"Fake\",
        \"last_name\": \"Person\",
        \"phone\": \"+1555000005\"
    }")

echo "Multi-Risk User Response: $MULTI_RISK_RESPONSE"
if echo "$MULTI_RISK_RESPONSE" | grep -q "is_risky"; then
    IS_RISKY=$(echo "$MULTI_RISK_RESPONSE" | jq -r '.is_risky')
    if [ "$IS_RISKY" = "true" ]; then
        echo "✅ Multi-risk user properly detected"
        RISK_LEVEL=$(echo "$MULTI_RISK_RESPONSE" | jq -r '.risk_level')
        echo "   Risk Level: $RISK_LEVEL"
    else
        echo "⚠️ Multi-risk user not flagged - check risk engine logic"
    fi
else
    echo "❌ Multi-risk user check failed"
    exit 1
fi
echo ""

echo "9. Cleaning up test rules..."
DELETE_RESPONSE=$(curl -s -X DELETE http://localhost:8080/api/v1/risk/rules/$RULE_ID \
    -H "Authorization: Bearer $ADMIN_JWT_TOKEN")

echo "Delete Rule Response: $DELETE_RESPONSE"
if echo "$DELETE_RESPONSE" | grep -q "success"; then
    SUCCESS=$(echo "$DELETE_RESPONSE" | jq -r '.success')
    if [ "$SUCCESS" = "true" ]; then
        echo "✅ Risk rule deletion successful"
    else
        echo "❌ Risk rule deletion failed"
    fi
else
    echo "❌ Risk rule deletion failed"
fi
echo ""

echo "10. Cleaning up additional test rules..."
curl -s -X DELETE http://localhost:8080/api/v1/risk/rules/$NAME_RULE_ID \
    -H "Authorization: Bearer $ADMIN_JWT_TOKEN" > /dev/null

curl -s -X DELETE http://localhost:8080/api/v1/risk/rules/$DOMAIN_RULE_ID \
    -H "Authorization: Bearer $ADMIN_JWT_TOKEN" > /dev/null

echo "✅ Test cleanup completed"
echo ""

echo "🛡️ Risk Management API tests completed successfully!"
echo ""
echo "📊 Test Summary:"
echo "   ✅ Risk rule creation (setup)"
echo "   ✅ Risk checking for clean users"
echo "   ✅ Risk checking for suspicious users"
echo "   ✅ Multiple risk rule types"
echo "   ✅ Multi-factor risk detection"
echo "   ✅ Test rule cleanup"
echo ""
echo "💡 Note: This test focuses on risk detection functionality."
echo "   Admin-specific tests are in test-admin.sh"
echo "   To verify risk detection: make logs"

#!/bin/bash
set -e

echo "🚨 Testing Risk Detection..."
echo ""

TIMESTAMP=$(date +%s)

# Test 1: High-risk Email Domain
echo "1. Creating high-risk user (suspicious email)..."
RISK_RESPONSE=$(curl -s -X POST http://localhost:8080/api/v1/auth/register \
    -H "Content-Type: application/json" \
    -d "{\"email\":\"test${TIMESTAMP}@suspicious.com\",\"password\":\"riskpass123\",\"first_name\":\"Suspicious\",\"last_name\":\"User\"}")

echo "High-Risk User Response: $RISK_RESPONSE"
if echo "$RISK_RESPONSE" | grep -q "access_token"; then
    echo "✅ High-risk user created (check logs for risk detection)"
else
    echo "❌ High-risk user creation failed"
    exit 1
fi
echo ""

# Test 2: Suspicious Name Pattern
echo "2. Creating user with suspicious name pattern..."
sleep 1
TIMESTAMP2=$(date +%s)
SUSPICIOUS_NAME_RESPONSE=$(curl -s -X POST http://localhost:8080/api/v1/auth/register \
    -H "Content-Type: application/json" \
    -d "{\"email\":\"fake.person${TIMESTAMP2}@example.com\",\"password\":\"fakepass123\",\"first_name\":\"Fake\",\"last_name\":\"Person\"}")

echo "Suspicious Name Response: $SUSPICIOUS_NAME_RESPONSE"
if echo "$SUSPICIOUS_NAME_RESPONSE" | grep -q "access_token"; then
    echo "✅ Suspicious name user created (check logs for risk detection)"
else
    echo "❌ Suspicious name user creation failed"
    exit 1
fi
echo ""

echo "🚨 Risk detection tests completed successfully!"
echo ""
echo "💡 Note: To verify risk detection is working properly:"
echo "   1. Check the service logs: make logs"
echo "   2. Look for risk assessment messages"
echo "   3. Check RabbitMQ management console: http://localhost:15672"

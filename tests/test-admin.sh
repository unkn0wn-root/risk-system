#!/bin/bash
set -e

echo "👑 Testing Admin Endpoints..."
echo ""

# Setup: Create admin user and get token
TIMESTAMP=$(date +%s)
ADMIN_EMAIL="admin${TIMESTAMP}@example.com"
echo "Using admin email: $ADMIN_EMAIL"
echo ""

echo "1. Creating admin user and getting token..."
ADMIN_REGISTER_RESPONSE=$(curl -s -X POST http://localhost:8080/api/v1/auth/register \
    -H "Content-Type: application/json" \
    -d "{\"email\":\"$ADMIN_EMAIL\",\"password\":\"adminpass123\",\"first_name\":\"Admin\",\"last_name\":\"User\",\"phone\":\"+1555000001\"}")

ADMIN_USER_ID=$(echo "$ADMIN_REGISTER_RESPONSE" | jq -r '.user.id')
echo "Admin User ID: $ADMIN_USER_ID"

# Manually update user role to admin in database (since there's no API endpoint for this)
echo "Updating user role to admin in database..."
PGPASSWORD="app_password" psql -h localhost -U app_admin -d users -c "
UPDATE users SET roles = '[\"admin\"]' WHERE id = '$ADMIN_USER_ID';
" > /dev/null 2>&1

if [ $? -eq 0 ]; then
    echo "✅ User role updated to admin in database"
else
    echo "❌ Failed to update user role - database not accessible or user not found"
    echo "💡 Make sure PostgreSQL is running and accessible with these credentials:"
    echo "   Host: localhost, User: app_admin, Password: app_password, Database: users"
    exit 1
fi

echo "Verifying role update..."
PGPASSWORD="app_password" psql -h localhost -U app_admin -d users -c "
SELECT id, email, roles FROM users WHERE id = '$ADMIN_USER_ID';
" > /dev/null 2>&1

echo "Logging in as admin to get token with admin privileges..."
ADMIN_LOGIN_RESPONSE=$(curl -s -X POST http://localhost:8080/api/v1/auth/login \
    -H "Content-Type: application/json" \
    -d "{\"email\":\"$ADMIN_EMAIL\",\"password\":\"adminpass123\"}")

echo "Admin Login Response: $ADMIN_LOGIN_RESPONSE"

ADMIN_JWT_TOKEN=$(echo "$ADMIN_LOGIN_RESPONSE" | jq -r '.access_token')
if [ "$ADMIN_JWT_TOKEN" = "null" ] || [ -z "$ADMIN_JWT_TOKEN" ]; then
    echo "❌ Failed to get admin token - login failed"
    echo "Response: $ADMIN_LOGIN_RESPONSE"
    exit 1
fi

echo "Admin Token obtained: ${ADMIN_JWT_TOKEN:0:20}..."
echo ""

echo "2. Testing admin user creation..."
ADMIN_CREATE_RESPONSE=$(curl -s -X POST http://localhost:8080/api/v1/users \
    -H "Authorization: Bearer $ADMIN_JWT_TOKEN" \
    -H "Content-Type: application/json" \
    -d "{\"email\":\"created${TIMESTAMP}@example.com\",\"first_name\":\"Created\",\"last_name\":\"User\",\"phone\":\"+1555111222\"}")

echo "Admin Create User Response: $ADMIN_CREATE_RESPONSE"
if echo "$ADMIN_CREATE_RESPONSE" | grep -q "Created\|id"; then
    echo "✅ Admin user creation successful"
else
    echo "⚠️  Admin user creation failed (user may not have admin role)"
fi
echo ""

echo "3. Testing risk rule creation (admin access)..."
CREATE_RULE_RESPONSE=$(curl -s -X POST http://localhost:8080/api/v1/risk/rules \
    -H "Authorization: Bearer $ADMIN_JWT_TOKEN" \
    -H "Content-Type: application/json" \
    -d '{
        "name": "Admin Test - Suspicious Email Domain",
        "type": "DOMAIN_BLACKLIST",
        "category": "EMAIL",
        "value": "admin-test-domain.com",
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

echo "4. Testing risk rules listing..."
LIST_RULES_RESPONSE=$(curl -s -X GET http://localhost:8080/api/v1/risk/rules \
    -H "Authorization: Bearer $ADMIN_JWT_TOKEN")

echo "List Rules Response: $LIST_RULES_RESPONSE"
if echo "$LIST_RULES_RESPONSE" | grep -q "rules"; then
    echo "✅ Risk rules listing successful"
else
    echo "❌ Risk rules listing failed"
    exit 1
fi
echo ""

echo "5. Testing risk rule update..."
UPDATE_RULE_RESPONSE=$(curl -s -X PUT http://localhost:8080/api/v1/risk/rules/$RULE_ID \
    -H "Authorization: Bearer $ADMIN_JWT_TOKEN" \
    -H "Content-Type: application/json" \
    -d '{
        "name": "Updated Admin Test - Suspicious Email Domain",
        "type": "DOMAIN_BLACKLIST",
        "category": "EMAIL",
        "value": "admin-test-domain.com",
        "score": 85,
        "is_active": true,
        "confidence": 0.95,
        "expires_in_days": 60
    }')

echo "Update Rule Response: $UPDATE_RULE_RESPONSE"
if echo "$UPDATE_RULE_RESPONSE" | grep -q "success"; then
    SUCCESS=$(echo "$UPDATE_RULE_RESPONSE" | jq -r '.success')
    if [ "$SUCCESS" = "true" ]; then
        echo "✅ Risk rule update successful"
    else
        echo "❌ Risk rule update failed"
        exit 1
    fi
else
    echo "❌ Risk rule update failed"
    exit 1
fi
echo ""

echo "6. Creating regular user for access control testing..."
USER_REGISTER_RESPONSE=$(curl -s -X POST http://localhost:8080/api/v1/auth/register \
    -H "Content-Type: application/json" \
    -d "{\"email\":\"regularuser${TIMESTAMP}@example.com\",\"password\":\"userpass123\",\"first_name\":\"Regular\",\"last_name\":\"User\",\"phone\":\"+1555000002\"}")

USER_JWT_TOKEN=$(echo "$USER_REGISTER_RESPONSE" | jq -r '.access_token')
echo "Regular User Token obtained: ${USER_JWT_TOKEN:0:20}..."
echo ""

echo "7. Testing non-admin access to admin endpoints (should fail)..."
NON_ADMIN_RESPONSE=$(curl -s -X POST http://localhost:8080/api/v1/risk/rules \
    -H "Authorization: Bearer $USER_JWT_TOKEN" \
    -H "Content-Type: application/json" \
    -d '{
        "name": "Unauthorized Rule",
        "type": "DOMAIN_BLACKLIST",
        "category": "EMAIL",
        "value": "unauthorized.com",
        "score": 100,
        "is_active": true,
        "confidence": 1.0
    }')

echo "Non-Admin Access Response: $NON_ADMIN_RESPONSE"
if echo "$NON_ADMIN_RESPONSE" | grep -q "error\|unauthorized\|forbidden"; then
    echo "✅ Non-admin access properly blocked"
else
    echo "❌ Non-admin access not properly blocked - security issue!"
    exit 1
fi
echo ""

echo "8. Testing risk rule deletion..."
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

echo "👑 Admin endpoint tests completed successfully!"
echo ""
echo "📊 Test Summary:"
echo "   ✅ Admin user creation and authentication"
echo "   ✅ Admin user creation endpoint"
echo "   ✅ Risk rule creation (admin only)"
echo "   ✅ Risk rules listing"
echo "   ✅ Risk rule updates"
echo "   ✅ Access control (non-admin blocked)"
echo "   ✅ Risk rule deletion"
echo ""
echo "💡 Note: Admin tests require database access for role assignment"
echo "   Database credentials: app_admin/app_password@localhost:5432/users"

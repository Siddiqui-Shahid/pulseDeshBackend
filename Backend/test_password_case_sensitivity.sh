#!/bin/bash

# Test the password case-sensitivity fix
# The backend should now properly reject pass123 when the correct password is Pass123

echo "Testing password case-sensitivity fix..."
echo ""

API_BASE="http://localhost:3000"

# Step 1: Create a user with password Pass123
echo "1. Creating user with password 'Pass123'..."
SIGNUP_RESPONSE=$(curl -s -X POST $API_BASE/auth/signup \
  -H "Content-Type: application/json" \
  -d '{
    "username": "testuser",
    "email": "test@example.com",
    "password": "Pass123"
  }')

echo "Response: $SIGNUP_RESPONSE"
echo ""

# Step 2: Try logging in with correct password (Pass123)
echo "2. Attempting login with CORRECT password 'Pass123'..."
LOGIN_CORRECT=$(curl -s -X POST $API_BASE/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test@example.com",
    "password": "Pass123"
  }')

if echo "$LOGIN_CORRECT" | grep -q '"success":true'; then
  echo "✅ LOGIN SUCCESSFUL with correct password (Pass123)"
else
  echo "❌ LOGIN FAILED with correct password - this is a problem!"
  echo "Response: $LOGIN_CORRECT"
fi
echo ""

# Step 3: Try logging in with incorrect lowercase password (pass123)
echo "3. Attempting login with INCORRECT password 'pass123'..."
LOGIN_WRONG=$(curl -s -X POST $API_BASE/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test@example.com",
    "password": "pass123"
  }')

if echo "$LOGIN_WRONG" | grep -q '"success":false'; then
  echo "✅ LOGIN REJECTED with incorrect password (pass123) - FIX CONFIRMED!"
else
  echo "❌ LOGIN ACCEPTED with incorrect password - BUG STILL EXISTS!"
  echo "Response: $LOGIN_WRONG"
fi
echo ""
echo "Test complete!"

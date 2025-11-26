#!/bin/bash

echo "Testing Wallet Creation UX Flow..."
echo

# Step 1: Register new user
echo "Step 1: Registering new test user..."
REGISTER_RESPONSE=$(curl -s -X POST http://localhost:8000/api/v1/identity/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "full_name": "UX Test User",
    "email": "uxtest@example.com",
    "phone": "9876543211",
    "password": "Test1234"
  }')

echo "$REGISTER_RESPONSE" | jq .
echo

# Step 2: Login
echo "Step 2: Logging in..."
LOGIN_RESPONSE=$(curl -s -X POST http://localhost:8000/api/v1/identity/auth/login \
  -H "Content-Type: application/json" \
  -d '{"identifier": "uxtest@example.com", "password": "Test1234"}')

TOKEN=$(echo "$LOGIN_RESPONSE" | jq -r '.data.token')
USER_ID=$(echo "$LOGIN_RESPONSE" | jq -r '.data.user.id')

echo "✅ Logged in with token"
echo

# Step 3: Check wallets (should be empty)
echo "Step 3: Checking wallets (should be empty)..."
WALLETS_RESPONSE=$(curl -s -X GET "http://localhost:8000/api/v1/wallet/wallets?user_id=$USER_ID" \
  -H "Authorization: Bearer $TOKEN")

WALLET_COUNT=$(echo "$WALLETS_RESPONSE" | jq '.data | length')
echo "Wallet count: $WALLET_COUNT"

if [ "$WALLET_COUNT" = "0" ]; then
  echo "✅ No wallets exist - dashboard should show 'Create Your First Wallet'"
else
  echo "❌ Wallets already exist"
fi
echo

# Step 4: Create wallet (simulating frontend button click)
echo "Step 4: Creating savings wallet..."
CREATE_RESPONSE=$(curl -s -X POST http://localhost:8000/api/v1/wallet/wallets \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d "{\"user_id\": \"$USER_ID\", \"type\": \"savings\", \"currency\": \"INR\"}")

echo "$CREATE_RESPONSE" | jq .

if echo "$CREATE_RESPONSE" | jq -e '.success' > /dev/null; then
  echo
  echo "✅ Wallet created successfully!"
  
  # Step 5: Verify wallet appears in list
  echo
  echo "Step 5: Verifying wallet appears in list..."
  WALLETS_AFTER=$(curl -s -X GET "http://localhost:8000/api/v1/wallet/wallets?user_id=$USER_ID" \
    -H "Authorization: Bearer $TOKEN")
  
  echo "$WALLETS_AFTER" | jq '.data[] | {id, type, currency, status, balance}'
  echo
  echo "✅ UX Flow Complete!"
else
  echo
  echo "❌ Wallet creation failed"
  echo "$CREATE_RESPONSE" | jq '.error'
fi

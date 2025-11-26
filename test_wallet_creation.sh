#!/bin/bash

# Test Wallet Creation Flow
echo "Testing Wallet Creation Flow..."
echo ""

# Step 1: Login
echo "Step 1: Logging in..."
LOGIN_RESPONSE=$(curl -s -X POST http://localhost:8000/api/v1/identity/auth/login \
  -H "Content-Type: application/json" \
  -d '{"identifier": "wallet99test@example.com", "password": "Test1234"}')

echo "$LOGIN_RESPONSE" | jq .

TOKEN=$(echo "$LOGIN_RESPONSE" | jq -r '.data.token')
USER_ID=$(echo "$LOGIN_RESPONSE" | jq -r '.data.user.id')

echo ""
echo "Token: ${TOKEN:0:50}..."
echo "User ID: $USER_ID"
echo ""

# Step 2: Create Wallet
echo "Step 2: Creating savings wallet..."
CREATE_RESPONSE=$(curl -s -X POST http://localhost:8000/api/v1/wallet/wallets \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d "{\"user_id\": \"$USER_ID\", \"type\": \"savings\", \"currency\": \"INR\"}")

echo "$CREATE_RESPONSE" | jq .

WALLET_ID=$(echo "$CREATE_RESPONSE" | jq -r '.data.id // empty')
echo ""

if [ -n "$WALLET_ID" ]; then
  echo "✅ Wallet created successfully! Wallet ID: $WALLET_ID"

  # Step 3: Verify wallet in database
  echo ""
  echo "Step 3: Verifying wallet in database..."
  docker exec nivo-postgres psql -U nivo -d nivo -c "SELECT id, user_id, type, currency, balance, status FROM wallets WHERE id = '$WALLET_ID';"

  # Step 4: Get wallets via API
  echo ""
  echo "Step 4: Fetching wallets via API..."
  GET_RESPONSE=$(curl -s -X GET http://localhost:8000/api/v1/wallet/wallets \
    -H "Authorization: Bearer $TOKEN")
  echo "$GET_RESPONSE" | jq .
else
  echo "❌ Wallet creation failed!"
  echo "$CREATE_RESPONSE" | jq '.error'
fi

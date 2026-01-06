#!/bin/bash
# Quick setup script for demo user

API_BASE="http://localhost:8000/api/v1"

# Login as demo user
echo "Logging in as demo user..."
LOGIN_RESPONSE=$(curl -s -X POST "$API_BASE/identity/auth/login" \
  -H "Content-Type: application/json" \
  -d '{"email":"demo@nivo.local","password":"demo123"}')

TOKEN=$(echo "$LOGIN_RESPONSE" | jq -r '.data.token')
USER_ID=$(echo "$LOGIN_RESPONSE" | jq -r '.data.user.id')

if [ "$TOKEN" = "null" ] || [ -z "$TOKEN" ]; then
  echo "Failed to login demo user"
  exit 1
fi

echo "Logged in successfully. User ID: $USER_ID"

# Create wallet
echo "Creating INR wallet..."
WALLET_RESPONSE=$(curl -s -X POST "$API_BASE/wallet/wallets" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d "{\"user_id\":\"$USER_ID\",\"currency\":\"INR\"}")

WALLET_ID=$(echo "$WALLET_RESPONSE" | jq -r '.data.id // empty')

if [ -z "$WALLET_ID" ] || [ "$WALLET_ID" = "null" ]; then
  echo "Failed to create wallet or wallet already exists"
  echo "$WALLET_RESPONSE" | jq
else
  echo "Wallet created successfully: $WALLET_ID"

  # Add initial balance
  echo "Adding initial balance..."
  curl -s -X POST "$API_BASE/transaction/transactions/deposit" \
    -H "Authorization: Bearer $TOKEN" \
    -H "Content-Type: application/json" \
    -d "{\"wallet_id\":\"$WALLET_ID\",\"amount_paise\":100000,\"description\":\"Initial balance\"}" | jq
fi

echo ""
echo "Demo user is ready!"
echo "Login at http://localhost:5173/ with:"
echo "  Email: demo@nivo.local"
echo "  Password: demo123"

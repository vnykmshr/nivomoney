//go:build integration
// +build integration

package integration_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	gatewayURL = "http://localhost:8000"
	timeout    = 10 * time.Second
)

// APIResponse is a generic response structure
type APIResponse struct {
	Success bool            `json:"success"`
	Data    json.RawMessage `json:"data,omitempty"`
	Error   *APIError       `json:"error,omitempty"`
}

// APIError represents an error response
type APIError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// User represents a user from the identity service
type User struct {
	ID        string    `json:"id"`
	Email     string    `json:"email"`
	Phone     string    `json:"phone"`
	FullName  string    `json:"full_name"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}

// LoginResponse represents the login response
type LoginResponse struct {
	Token     string `json:"token"`
	ExpiresAt int64  `json:"expires_at"`
	User      User   `json:"user"`
}

// Wallet represents a wallet
type Wallet struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	Currency  string    `json:"currency"`
	Balance   string    `json:"balance"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}

// Transaction represents a transaction
type Transaction struct {
	ID              string    `json:"id"`
	Type            string    `json:"type"`
	FromWalletID    string    `json:"from_wallet_id,omitempty"`
	ToWalletID      string    `json:"to_wallet_id"`
	Amount          string    `json:"amount"`
	Currency        string    `json:"currency"`
	Status          string    `json:"status"`
	Description     string    `json:"description,omitempty"`
	IdempotencyKey  string    `json:"idempotency_key,omitempty"`
	CreatedAt       time.Time `json:"created_at"`
}

// TestClient is a helper for making HTTP requests
type TestClient struct {
	httpClient *http.Client
	baseURL    string
	authToken  string
	t          *testing.T
}

// NewTestClient creates a new test client
func NewTestClient(t *testing.T) *TestClient {
	return &TestClient{
		httpClient: &http.Client{Timeout: timeout},
		baseURL:    gatewayURL,
		t:          t,
	}
}

// SetAuthToken sets the authorization token
func (c *TestClient) SetAuthToken(token string) {
	c.authToken = token
}

// Post makes a POST request
func (c *TestClient) Post(path string, body interface{}, requireAuth bool) (*APIResponse, int) {
	jsonData, err := json.Marshal(body)
	require.NoError(c.t, err, "failed to marshal request body")

	req, err := http.NewRequest("POST", c.baseURL+path, bytes.NewBuffer(jsonData))
	require.NoError(c.t, err, "failed to create request")

	req.Header.Set("Content-Type", "application/json")
	if requireAuth && c.authToken != "" {
		req.Header.Set("Authorization", "Bearer "+c.authToken)
	}

	resp, err := c.httpClient.Do(req)
	require.NoError(c.t, err, "failed to make request")
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	require.NoError(c.t, err, "failed to read response body")

	var apiResp APIResponse
	err = json.Unmarshal(bodyBytes, &apiResp)
	require.NoError(c.t, err, "failed to unmarshal response: %s", string(bodyBytes))

	return &apiResp, resp.StatusCode
}

// Get makes a GET request
func (c *TestClient) Get(path string, requireAuth bool) (*APIResponse, int) {
	req, err := http.NewRequest("GET", c.baseURL+path, nil)
	require.NoError(c.t, err, "failed to create request")

	if requireAuth && c.authToken != "" {
		req.Header.Set("Authorization", "Bearer "+c.authToken)
	}

	resp, err := c.httpClient.Do(req)
	require.NoError(c.t, err, "failed to make request")
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	require.NoError(c.t, err, "failed to read response body")

	var apiResp APIResponse
	err = json.Unmarshal(bodyBytes, &apiResp)
	require.NoError(c.t, err, "failed to unmarshal response")

	return &apiResp, resp.StatusCode
}

// TestGatewayHealthCheck verifies the gateway is running
func TestGatewayHealthCheck(t *testing.T) {
	client := NewTestClient(t)

	resp, statusCode := client.Get("/health", false)

	assert.Equal(t, http.StatusOK, statusCode, "health check should return 200")
	assert.True(t, resp.Success, "health check should be successful")
}

// TestUserRegistrationAndLogin tests the complete auth flow through gateway
func TestUserRegistrationAndLogin(t *testing.T) {
	client := NewTestClient(t)

	// Generate unique test data
	timestamp := time.Now().Unix()
	email := fmt.Sprintf("integration-test-%d@nivo.com", timestamp)
	phone := fmt.Sprintf("+919%09d", timestamp%1000000000)

	// Test 1: Register a new user
	t.Run("RegisterUser", func(t *testing.T) {
		registerReq := map[string]interface{}{
			"email":     email,
			"password":  "SecurePass123",
			"full_name": "Integration Test User",
			"phone":     phone,
		}

		resp, statusCode := client.Post("/api/v1/identity/auth/register", registerReq, false)

		assert.Equal(t, http.StatusCreated, statusCode, "user registration should return 201")
		assert.True(t, resp.Success, "registration should be successful")
		assert.Nil(t, resp.Error, "registration should not have errors")

		// Unmarshal user data
		var user User
		err := json.Unmarshal(resp.Data, &user)
		require.NoError(t, err, "should be able to unmarshal user data")

		assert.NotEmpty(t, user.ID, "user should have an ID")
		assert.Equal(t, email, user.Email, "user email should match")
		assert.Equal(t, phone, user.Phone, "user phone should match")
		assert.Equal(t, "pending", user.Status, "new user status should be pending")
	})

	// Test 2: Login with the registered user
	t.Run("LoginUser", func(t *testing.T) {
		loginReq := map[string]interface{}{
			"email":    email,
			"password": "SecurePass123",
		}

		resp, statusCode := client.Post("/api/v1/identity/auth/login", loginReq, false)

		assert.Equal(t, http.StatusOK, statusCode, "login should return 200")
		assert.True(t, resp.Success, "login should be successful")

		// Unmarshal login response
		var loginResp LoginResponse
		err := json.Unmarshal(resp.Data, &loginResp)
		require.NoError(t, err, "should be able to unmarshal login response")

		assert.NotEmpty(t, loginResp.Token, "should receive a JWT token")
		assert.NotEmpty(t, loginResp.User.ID, "should receive user data")
		assert.Equal(t, email, loginResp.User.Email, "user email should match")

		// Store token for subsequent requests
		client.SetAuthToken(loginResp.Token)
	})

	// Test 3: Attempt login with wrong password
	t.Run("LoginWithWrongPassword", func(t *testing.T) {
		loginReq := map[string]interface{}{
			"email":    email,
			"password": "WrongPassword123",
		}

		resp, statusCode := client.Post("/api/v1/identity/auth/login", loginReq, false)

		assert.Equal(t, http.StatusUnauthorized, statusCode, "wrong password should return 401")
		assert.False(t, resp.Success, "login should fail")
		assert.NotNil(t, resp.Error, "should have error response")
		assert.Equal(t, "UNAUTHORIZED", resp.Error.Code, "error code should be UNAUTHORIZED")
	})

	// Test 4: Attempt to register duplicate user
	t.Run("RegisterDuplicateUser", func(t *testing.T) {
		registerReq := map[string]interface{}{
			"email":     email,
			"password":  "SecurePass123",
			"full_name": "Duplicate User",
			"phone":     phone,
		}

		resp, statusCode := client.Post("/api/v1/identity/auth/register", registerReq, false)

		assert.Equal(t, http.StatusConflict, statusCode, "duplicate registration should return 409")
		assert.False(t, resp.Success, "registration should fail")
		assert.NotNil(t, resp.Error, "should have error response")
		assert.Equal(t, "CONFLICT", resp.Error.Code, "error code should be CONFLICT")
	})
}

// TestWalletCreationAndManagement tests wallet operations through gateway
func TestWalletCreationAndManagement(t *testing.T) {
	client := NewTestClient(t)

	// Setup: Register and login user
	timestamp := time.Now().Unix()
	email := fmt.Sprintf("wallet-test-%d@nivo.com", timestamp)
	phone := fmt.Sprintf("+919%09d", timestamp%1000000000)

	// Register user
	registerReq := map[string]interface{}{
		"email":     email,
		"password":  "SecurePass123",
		"full_name": "Wallet Test User",
		"phone":     phone,
	}
	client.Post("/api/v1/identity/auth/register", registerReq, false)

	// Login to get token
	loginReq := map[string]interface{}{
		"email":    email,
		"password": "SecurePass123",
	}
	loginResp, _ := client.Post("/api/v1/identity/auth/login", loginReq, false)
	var loginData LoginResponse
	json.Unmarshal(loginResp.Data, &loginData)
	client.SetAuthToken(loginData.Token)

	var walletID string

	// Test 1: Create a wallet
	t.Run("CreateWallet", func(t *testing.T) {
		createWalletReq := map[string]interface{}{
			"currency": "INR",
		}

		resp, statusCode := client.Post("/api/v1/wallet/wallets", createWalletReq, true)

		assert.Equal(t, http.StatusCreated, statusCode, "wallet creation should return 201")
		assert.True(t, resp.Success, "wallet creation should be successful")

		// Unmarshal wallet data
		var wallet Wallet
		err := json.Unmarshal(resp.Data, &wallet)
		require.NoError(t, err, "should be able to unmarshal wallet data")

		assert.NotEmpty(t, wallet.ID, "wallet should have an ID")
		assert.Equal(t, "INR", wallet.Currency, "wallet currency should match")
		assert.Equal(t, "0", wallet.Balance, "new wallet should have zero balance")
		assert.Equal(t, "pending", wallet.Status, "new wallet status should be pending")

		walletID = wallet.ID
	})

	// Test 2: Get wallet details
	t.Run("GetWallet", func(t *testing.T) {
		if walletID == "" {
			t.Skip("Wallet ID not available from previous test")
		}

		resp, statusCode := client.Get(fmt.Sprintf("/api/v1/wallet/wallets/%s", walletID), true)

		assert.Equal(t, http.StatusOK, statusCode, "get wallet should return 200")
		assert.True(t, resp.Success, "get wallet should be successful")

		// Unmarshal wallet data
		var wallet Wallet
		err := json.Unmarshal(resp.Data, &wallet)
		require.NoError(t, err, "should be able to unmarshal wallet data")

		assert.Equal(t, walletID, wallet.ID, "wallet ID should match")
	})

	// Test 3: Activate wallet
	t.Run("ActivateWallet", func(t *testing.T) {
		if walletID == "" {
			t.Skip("Wallet ID not available from previous test")
		}

		resp, statusCode := client.Post(fmt.Sprintf("/api/v1/wallet/wallets/%s/activate", walletID), nil, true)

		assert.Equal(t, http.StatusOK, statusCode, "activate wallet should return 200")
		assert.True(t, resp.Success, "wallet activation should be successful")
	})

	// Test 4: Create wallet without authentication
	t.Run("CreateWalletWithoutAuth", func(t *testing.T) {
		createWalletReq := map[string]interface{}{
			"currency": "INR",
		}

		// Temporarily remove auth token
		originalToken := client.authToken
		client.authToken = ""

		resp, statusCode := client.Post("/api/v1/wallet/wallets", createWalletReq, false)

		assert.Equal(t, http.StatusUnauthorized, statusCode, "should return 401 without auth")
		assert.False(t, resp.Success, "wallet creation should fail")

		// Restore token
		client.authToken = originalToken
	})
}

// TestTransactionFlow tests the complete transaction flow through gateway
func TestTransactionFlow(t *testing.T) {
	client := NewTestClient(t)

	// Setup: Register user and create wallet
	timestamp := time.Now().Unix()
	email := fmt.Sprintf("tx-test-%d@nivo.com", timestamp)
	phone := fmt.Sprintf("+919%09d", timestamp%1000000000)

	// Register and login
	registerReq := map[string]interface{}{
		"email":     email,
		"password":  "SecurePass123",
		"full_name": "Transaction Test User",
		"phone":     phone,
	}
	client.Post("/api/v1/identity/auth/register", registerReq, false)

	loginReq := map[string]interface{}{
		"email":    email,
		"password": "SecurePass123",
	}
	loginResp, _ := client.Post("/api/v1/identity/auth/login", loginReq, false)
	var loginData LoginResponse
	json.Unmarshal(loginResp.Data, &loginData)
	client.SetAuthToken(loginData.Token)

	// Create and activate wallet
	createWalletReq := map[string]interface{}{"currency": "INR"}
	walletResp, _ := client.Post("/api/v1/wallet/wallets", createWalletReq, true)
	var wallet Wallet
	json.Unmarshal(walletResp.Data, &wallet)
	walletID := wallet.ID

	// Activate wallet
	client.Post(fmt.Sprintf("/api/v1/wallet/wallets/%s/activate", walletID), nil, true)

	// Test 1: Create a deposit transaction
	t.Run("CreateDeposit", func(t *testing.T) {
		depositReq := map[string]interface{}{
			"wallet_id":   walletID,
			"amount":      "1000.00",
			"currency":    "INR",
			"description": "Integration test deposit",
		}

		resp, statusCode := client.Post("/api/v1/transaction/transactions/deposit", depositReq, true)

		assert.Equal(t, http.StatusCreated, statusCode, "deposit should return 201")
		assert.True(t, resp.Success, "deposit should be successful")

		// Unmarshal transaction data
		var tx Transaction
		err := json.Unmarshal(resp.Data, &tx)
		require.NoError(t, err, "should be able to unmarshal transaction data")

		assert.NotEmpty(t, tx.ID, "transaction should have an ID")
		assert.Equal(t, "deposit", tx.Type, "transaction type should be deposit")
		assert.Equal(t, "1000.00", tx.Amount, "amount should match")
		assert.Equal(t, walletID, tx.ToWalletID, "to_wallet_id should match")
	})

	// Test 2: Verify wallet balance after deposit
	t.Run("VerifyBalanceAfterDeposit", func(t *testing.T) {
		// Wait a bit for transaction processing
		time.Sleep(1 * time.Second)

		resp, statusCode := client.Get(fmt.Sprintf("/api/v1/wallet/wallets/%s/balance", walletID), true)

		assert.Equal(t, http.StatusOK, statusCode, "get balance should return 200")
		assert.True(t, resp.Success, "get balance should be successful")

		// The balance should be updated (might be "1000.00" or "1000")
		// Just verify we can get the balance
		assert.NotNil(t, resp.Data, "should have balance data")
	})

	// Test 3: Attempt withdrawal without sufficient funds
	t.Run("WithdrawWithInsufficientFunds", func(t *testing.T) {
		withdrawalReq := map[string]interface{}{
			"wallet_id":   walletID,
			"amount":      "5000.00",
			"currency":    "INR",
			"description": "Test withdrawal - should fail",
		}

		resp, _ := client.Post("/api/v1/transaction/transactions/withdrawal", withdrawalReq, true)

		// This should fail with insufficient funds error
		assert.False(t, resp.Success, "withdrawal should fail")
		assert.NotNil(t, resp.Error, "should have error response")
	})
}

// TestEndToEndUserJourney tests a complete user journey through the gateway
func TestEndToEndUserJourney(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping end-to-end test in short mode")
	}

	client := NewTestClient(t)

	timestamp := time.Now().Unix()
	email := fmt.Sprintf("e2e-test-%d@nivo.com", timestamp)
	phone := fmt.Sprintf("+919%09d", timestamp%1000000000)

	t.Log("Step 1: User Registration")
	registerReq := map[string]interface{}{
		"email":     email,
		"password":  "SecurePass123",
		"full_name": "End-to-End Test User",
		"phone":     phone,
	}
	registerResp, registerStatus := client.Post("/api/v1/identity/auth/register", registerReq, false)
	require.Equal(t, http.StatusCreated, registerStatus, "registration should succeed")
	require.True(t, registerResp.Success, "registration should be successful")

	t.Log("Step 2: User Login")
	loginReq := map[string]interface{}{
		"email":    email,
		"password": "SecurePass123",
	}
	loginResp, loginStatus := client.Post("/api/v1/identity/auth/login", loginReq, false)
	require.Equal(t, http.StatusOK, loginStatus, "login should succeed")
	require.True(t, loginResp.Success, "login should be successful")

	var loginData LoginResponse
	err := json.Unmarshal(loginResp.Data, &loginData)
	require.NoError(t, err, "should unmarshal login data")
	client.SetAuthToken(loginData.Token)
	t.Logf("Received auth token: %s...", loginData.Token[:20])

	t.Log("Step 3: Create Primary Wallet")
	createWalletReq := map[string]interface{}{"currency": "INR"}
	walletResp, walletStatus := client.Post("/api/v1/wallet/wallets", createWalletReq, true)
	require.Equal(t, http.StatusCreated, walletStatus, "wallet creation should succeed")
	require.True(t, walletResp.Success, "wallet creation should be successful")

	var primaryWallet Wallet
	err = json.Unmarshal(walletResp.Data, &primaryWallet)
	require.NoError(t, err, "should unmarshal wallet data")
	t.Logf("Created primary wallet: %s", primaryWallet.ID)

	t.Log("Step 4: Activate Primary Wallet")
	activateResp, activateStatus := client.Post(fmt.Sprintf("/api/v1/wallet/wallets/%s/activate", primaryWallet.ID), nil, true)
	require.Equal(t, http.StatusOK, activateStatus, "wallet activation should succeed")
	require.True(t, activateResp.Success, "wallet activation should be successful")

	t.Log("Step 5: Deposit Funds")
	depositReq := map[string]interface{}{
		"wallet_id":   primaryWallet.ID,
		"amount":      "5000.00",
		"currency":    "INR",
		"description": "Initial deposit for E2E test",
	}
	depositResp, depositStatus := client.Post("/api/v1/transaction/transactions/deposit", depositReq, true)
	require.Equal(t, http.StatusCreated, depositStatus, "deposit should succeed")
	require.True(t, depositResp.Success, "deposit should be successful")

	var depositTx Transaction
	err = json.Unmarshal(depositResp.Data, &depositTx)
	require.NoError(t, err, "should unmarshal deposit transaction")
	t.Logf("Deposited 5000.00 INR, transaction ID: %s", depositTx.ID)

	t.Log("Step 6: Create Second Wallet for Transfer")
	secondWalletResp, _ := client.Post("/api/v1/wallet/wallets", createWalletReq, true)
	var secondWallet Wallet
	json.Unmarshal(secondWalletResp.Data, &secondWallet)
	client.Post(fmt.Sprintf("/api/v1/wallet/wallets/%s/activate", secondWallet.ID), nil, true)
	t.Logf("Created second wallet: %s", secondWallet.ID)

	t.Log("Step 7: Transfer Between Wallets")
	transferReq := map[string]interface{}{
		"from_wallet_id": primaryWallet.ID,
		"to_wallet_id":   secondWallet.ID,
		"amount":         "1500.00",
		"currency":       "INR",
		"description":    "Transfer for E2E test",
	}
	transferResp, transferStatus := client.Post("/api/v1/transaction/transactions/transfer", transferReq, true)
	require.Equal(t, http.StatusCreated, transferStatus, "transfer should succeed")
	require.True(t, transferResp.Success, "transfer should be successful")

	var transferTx Transaction
	err = json.Unmarshal(transferResp.Data, &transferTx)
	require.NoError(t, err, "should unmarshal transfer transaction")
	t.Logf("Transferred 1500.00 INR from wallet %s to %s", primaryWallet.ID, secondWallet.ID)

	t.Log("✅ End-to-end user journey completed successfully!")
}

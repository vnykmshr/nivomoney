package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
)

// GatewayClient makes API calls to the Nivo Gateway
type GatewayClient struct {
	baseURL    string
	httpClient *http.Client
	authToken  string // Admin token for simulations
}

// NewGatewayClient creates a new gateway client
func NewGatewayClient(baseURL, authToken string) *GatewayClient {
	return &GatewayClient{
		baseURL:   baseURL,
		authToken: authToken,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// DepositRequest represents a deposit transaction request
type DepositRequest struct {
	WalletID    string `json:"wallet_id"`
	AmountPaise int64  `json:"amount_paise"`
	Description string `json:"description"`
}

// TransferRequest represents a transfer transaction request
type TransferRequest struct {
	SourceWalletID      string `json:"source_wallet_id"`
	DestinationWalletID string `json:"destination_wallet_id"`
	AmountPaise         int64  `json:"amount_paise"`
	Description         string `json:"description"`
}

// WithdrawalRequest represents a withdrawal transaction request
type WithdrawalRequest struct {
	WalletID    string `json:"wallet_id"`
	AmountPaise int64  `json:"amount_paise"`
	Description string `json:"description"`
}

// RegisterRequest represents a user registration request
type RegisterRequest struct {
	Email    string `json:"email"`
	Phone    string `json:"phone"`
	FullName string `json:"full_name"`
	Password string `json:"password"`
}

// RegisterResponse represents the response from user registration
type RegisterResponse struct {
	ID       string `json:"id"`
	Email    string `json:"email"`
	Phone    string `json:"phone"`
	FullName string `json:"full_name"`
	Status   string `json:"status"`
}

// LoginRequest represents a login request
type LoginRequest struct {
	Identifier string `json:"identifier"`
	Password   string `json:"password"`
}

// LoginResponse represents the response from login
type LoginResponse struct {
	Token string `json:"token"`
	User  struct {
		ID       string `json:"id"`
		Email    string `json:"email"`
		FullName string `json:"full_name"`
	} `json:"user"`
}

// KYCSubmitRequest represents a KYC submission request
type KYCSubmitRequest struct {
	FirstName    string `json:"first_name"`
	LastName     string `json:"last_name"`
	DateOfBirth  string `json:"date_of_birth"`
	PanNumber    string `json:"pan_number"`
	AadharNumber string `json:"aadhar_number"`
	Address      struct {
		Street  string `json:"street"`
		City    string `json:"city"`
		State   string `json:"state"`
		Pincode string `json:"pincode"`
		Country string `json:"country"`
	} `json:"address"`
}

// CreateDeposit creates a deposit transaction
func (c *GatewayClient) CreateDeposit(walletID string, amountPaise int64, description string) error {
	req := DepositRequest{
		WalletID:    walletID,
		AmountPaise: amountPaise,
		Description: description,
	}

	return c.makeRequest("POST", "/api/v1/transaction/transactions/deposit", req)
}

// CreateTransfer creates a transfer transaction
func (c *GatewayClient) CreateTransfer(sourceWalletID, destWalletID string, amountPaise int64, description string) error {
	req := TransferRequest{
		SourceWalletID:      sourceWalletID,
		DestinationWalletID: destWalletID,
		AmountPaise:         amountPaise,
		Description:         description,
	}

	return c.makeRequest("POST", "/api/v1/transaction/transactions/transfer", req)
}

// CreateWithdrawal creates a withdrawal transaction
func (c *GatewayClient) CreateWithdrawal(walletID string, amountPaise int64, description string) error {
	req := WithdrawalRequest{
		WalletID:    walletID,
		AmountPaise: amountPaise,
		Description: description,
	}

	return c.makeRequest("POST", "/api/v1/transaction/transactions/withdrawal", req)
}

// RegisterUser creates a new user account
func (c *GatewayClient) RegisterUser(email, phone, fullName, password string) (*RegisterResponse, error) {
	req := RegisterRequest{
		Email:    email,
		Phone:    phone,
		FullName: fullName,
		Password: password,
	}

	var resp RegisterResponse
	if err := c.makeRequestWithResponse("POST", "/api/v1/auth/register", req, &resp); err != nil {
		return nil, err
	}

	log.Printf("[simulation] ✓ User registered: %s", email)
	return &resp, nil
}

// Login authenticates a user and returns a session token
func (c *GatewayClient) Login(identifier, password string) (*LoginResponse, error) {
	req := LoginRequest{
		Identifier: identifier,
		Password:   password,
	}

	var resp LoginResponse
	if err := c.makeRequestWithResponse("POST", "/api/v1/auth/login", req, &resp); err != nil {
		return nil, err
	}

	log.Printf("[simulation] ✓ User logged in: %s", identifier)
	return &resp, nil
}

// Logout terminates a user session
func (c *GatewayClient) Logout(token string) error {
	// Create custom request with token
	url := c.baseURL + "/api/v1/auth/logout"
	req, err := http.NewRequest("POST", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode >= 400 {
		responseBody, _ := io.ReadAll(resp.Body)
		log.Printf("[simulation] Logout error %d: %s", resp.StatusCode, string(responseBody))
		return fmt.Errorf("logout failed with status %d", resp.StatusCode)
	}

	log.Printf("[simulation] ✓ User logged out")
	return nil
}

// SubmitKYC submits KYC information for a user
func (c *GatewayClient) SubmitKYC(token string, kycReq KYCSubmitRequest) error {
	// Create custom request with user token
	jsonData, err := json.Marshal(kycReq)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	url := c.baseURL + "/api/v1/kyc/submit"
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	responseBody, _ := io.ReadAll(resp.Body)

	if resp.StatusCode >= 400 {
		log.Printf("[simulation] KYC submission error %d: %s", resp.StatusCode, string(responseBody))
		return fmt.Errorf("KYC submission failed with status %d", resp.StatusCode)
	}

	log.Printf("[simulation] ✓ KYC submitted")
	return nil
}

// VerifyKYC admin endpoint to verify KYC (requires admin token)
func (c *GatewayClient) VerifyKYC(userID string) error {
	// This would be an admin endpoint - using the admin token from client
	url := fmt.Sprintf("%s/api/v1/admin/kyc/%s/verify", c.baseURL, userID)
	req, err := http.NewRequest("POST", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.authToken)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode >= 400 {
		responseBody, _ := io.ReadAll(resp.Body)
		log.Printf("[simulation] KYC verification error %d: %s", resp.StatusCode, string(responseBody))
		return fmt.Errorf("KYC verification failed with status %d", resp.StatusCode)
	}

	log.Printf("[simulation] ✓ KYC verified for user %s", userID)
	return nil
}

// WalletResponse represents the response from wallet query
type WalletResponse struct {
	ID       string `json:"id"`
	UserID   string `json:"user_id"`
	Currency string `json:"currency"`
	Balance  int64  `json:"balance"`
	Status   string `json:"status"`
}

// GetUserWallet fetches the wallet for a given user
func (c *GatewayClient) GetUserWallet(token, userID string) (*WalletResponse, error) {
	url := fmt.Sprintf("%s/api/v1/wallet/wallets/user/%s", c.baseURL, userID)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode >= 400 {
		log.Printf("[simulation] Get wallet error %d: %s", resp.StatusCode, string(responseBody))
		return nil, fmt.Errorf("get wallet failed with status %d", resp.StatusCode)
	}

	// The response might be a single wallet or an array - handle both
	// First try as array
	var wallets []WalletResponse
	if err := json.Unmarshal(responseBody, &wallets); err == nil && len(wallets) > 0 {
		// Return first wallet (usually the default INR wallet)
		return &wallets[0], nil
	}

	// Try as single wallet
	var wallet WalletResponse
	if err := json.Unmarshal(responseBody, &wallet); err != nil {
		return nil, fmt.Errorf("failed to decode wallet response: %w", err)
	}

	return &wallet, nil
}

// makeRequest is a helper to make HTTP requests
func (c *GatewayClient) makeRequest(method, path string, body interface{}) error {
	jsonData, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	url := c.baseURL + path
	req, err := http.NewRequest(method, url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	if c.authToken != "" {
		req.Header.Set("Authorization", "Bearer "+c.authToken)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	responseBody, _ := io.ReadAll(resp.Body)

	if resp.StatusCode >= 400 {
		log.Printf("[simulation] API error %d: %s", resp.StatusCode, string(responseBody))
		return fmt.Errorf("API returned status %d", resp.StatusCode)
	}

	log.Printf("[simulation] Transaction created successfully: %s %s", method, path)
	return nil
}

// makeRequestWithResponse makes an HTTP request and decodes the response
func (c *GatewayClient) makeRequestWithResponse(method, path string, body interface{}, response interface{}) error {
	jsonData, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	url := c.baseURL + path
	req, err := http.NewRequest(method, url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	if c.authToken != "" {
		req.Header.Set("Authorization", "Bearer "+c.authToken)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode >= 400 {
		log.Printf("[simulation] API error %d: %s", resp.StatusCode, string(responseBody))
		return fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(responseBody))
	}

	if err := json.Unmarshal(responseBody, response); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}

	return nil
}

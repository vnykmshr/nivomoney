package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/vnykmshr/nivo/shared/clients"
)

// GatewayClient makes API calls to the Nivo Gateway
type GatewayClient struct {
	*clients.BaseClient
}

// NewGatewayClient creates a new gateway client with admin auth token
func NewGatewayClient(baseURL, authToken string) *GatewayClient {
	var headers map[string]string
	if authToken != "" {
		headers = map[string]string{
			"Authorization": "Bearer " + authToken,
		}
	}
	return &GatewayClient{
		BaseClient: clients.NewBaseClientWithHeaders(baseURL, clients.DefaultTimeout, headers),
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
// Matches identity service's UpdateKYCRequest format
type KYCSubmitRequest struct {
	PAN         string            `json:"pan"`
	Aadhaar     string            `json:"aadhaar"`
	DateOfBirth string            `json:"date_of_birth"`
	Address     KYCAddressRequest `json:"address"`
}

// KYCAddressRequest represents address in KYC request
type KYCAddressRequest struct {
	Street  string `json:"street"`
	City    string `json:"city"`
	State   string `json:"state"`
	PIN     string `json:"pin"`
	Country string `json:"country"`
}

// WalletResponse represents the response from wallet query
type WalletResponse struct {
	ID       string `json:"id"`
	UserID   string `json:"user_id"`
	Currency string `json:"currency"`
	Balance  int64  `json:"balance"`
	Status   string `json:"status"`
}

// bearerToken creates auth headers for a given token.
func bearerToken(token string) map[string]string {
	return map[string]string{"Authorization": "Bearer " + token}
}

// CreateDeposit creates a deposit transaction.
// If token is provided, it's used for auth. Otherwise, falls back to client's default auth headers.
func (c *GatewayClient) CreateDeposit(ctx context.Context, token, walletID string, amountPaise int64, description string) error {
	req := DepositRequest{
		WalletID:    walletID,
		AmountPaise: amountPaise,
		Description: description,
	}

	var err error
	if token != "" {
		err = c.PostWithHeaders(ctx, "/api/v1/transaction/transactions/deposit", req, nil, bearerToken(token))
	} else {
		err = c.Post(ctx, "/api/v1/transaction/transactions/deposit", req, nil)
	}
	if err != nil {
		return err
	}
	log.Printf("[simulation] Transaction created successfully: POST /api/v1/transaction/transactions/deposit")
	return nil
}

// CreateTransfer creates a transfer transaction.
// If token is provided, it's used for auth. Otherwise, falls back to client's default auth headers.
func (c *GatewayClient) CreateTransfer(ctx context.Context, token, sourceWalletID, destWalletID string, amountPaise int64, description string) error {
	req := TransferRequest{
		SourceWalletID:      sourceWalletID,
		DestinationWalletID: destWalletID,
		AmountPaise:         amountPaise,
		Description:         description,
	}

	var err error
	if token != "" {
		err = c.PostWithHeaders(ctx, "/api/v1/transaction/transactions/transfer", req, nil, bearerToken(token))
	} else {
		err = c.Post(ctx, "/api/v1/transaction/transactions/transfer", req, nil)
	}
	if err != nil {
		return err
	}
	log.Printf("[simulation] Transaction created successfully: POST /api/v1/transaction/transactions/transfer")
	return nil
}

// CreateWithdrawal creates a withdrawal transaction.
// If token is provided, it's used for auth. Otherwise, falls back to client's default auth headers.
func (c *GatewayClient) CreateWithdrawal(ctx context.Context, token, walletID string, amountPaise int64, description string) error {
	req := WithdrawalRequest{
		WalletID:    walletID,
		AmountPaise: amountPaise,
		Description: description,
	}

	var err error
	if token != "" {
		err = c.PostWithHeaders(ctx, "/api/v1/transaction/transactions/withdrawal", req, nil, bearerToken(token))
	} else {
		err = c.Post(ctx, "/api/v1/transaction/transactions/withdrawal", req, nil)
	}
	if err != nil {
		return err
	}
	log.Printf("[simulation] Transaction created successfully: POST /api/v1/transaction/transactions/withdrawal")
	return nil
}

// RegisterUser creates a new user account
func (c *GatewayClient) RegisterUser(ctx context.Context, email, phone, fullName, password string) (*RegisterResponse, error) {
	req := RegisterRequest{
		Email:    email,
		Phone:    phone,
		FullName: fullName,
		Password: password,
	}

	var resp RegisterResponse
	if err := c.Post(ctx, "/api/v1/auth/register", req, &resp); err != nil {
		return nil, err
	}

	log.Printf("[simulation] ✓ User registered: %s", email)
	return &resp, nil
}

// Login authenticates a user and returns a session token
func (c *GatewayClient) Login(ctx context.Context, identifier, password string) (*LoginResponse, error) {
	req := LoginRequest{
		Identifier: identifier,
		Password:   password,
	}

	var resp LoginResponse
	if err := c.Post(ctx, "/api/v1/auth/login", req, &resp); err != nil {
		return nil, err
	}

	log.Printf("[simulation] ✓ User logged in: %s", identifier)
	return &resp, nil
}

// Logout terminates a user session
func (c *GatewayClient) Logout(ctx context.Context, token string) error {
	if err := c.PostWithHeaders(ctx, "/api/v1/auth/logout", nil, nil, bearerToken(token)); err != nil {
		return err
	}
	log.Printf("[simulation] ✓ User logged out")
	return nil
}

// SubmitKYC submits KYC information for a user
func (c *GatewayClient) SubmitKYC(ctx context.Context, token string, kycReq KYCSubmitRequest) error {
	// Route: /api/v1/identity/auth/kyc -> identity service's /api/v1/auth/kyc
	// Uses PUT method per identity service API
	if err := c.PutWithHeaders(ctx, "/api/v1/identity/auth/kyc", kycReq, nil, bearerToken(token)); err != nil {
		return err
	}
	log.Printf("[simulation] ✓ KYC submitted")
	return nil
}

// VerifyKYC admin endpoint to verify KYC (requires admin token)
func (c *GatewayClient) VerifyKYC(ctx context.Context, userID string) error {
	path := fmt.Sprintf("/api/v1/admin/kyc/%s/verify", userID)
	if err := c.Post(ctx, path, nil, nil); err != nil {
		return err
	}
	log.Printf("[simulation] ✓ KYC verified for user %s", userID)
	return nil
}

// GetUserWallet fetches the wallet for a given user
func (c *GatewayClient) GetUserWallet(ctx context.Context, token, userID string) (*WalletResponse, error) {
	// Route: /api/v1/wallet/users/:userID/wallets -> wallet service's /api/v1/users/:userID/wallets
	path := fmt.Sprintf("/api/v1/wallet/users/%s/wallets", userID)

	// This endpoint can return array or single wallet, so we parse as raw JSON first
	var rawResponse json.RawMessage
	if err := c.GetWithHeaders(ctx, path, &rawResponse, bearerToken(token)); err != nil {
		return nil, err
	}

	// Try as array first
	var wallets []WalletResponse
	if err := json.Unmarshal(rawResponse, &wallets); err == nil && len(wallets) > 0 {
		return &wallets[0], nil
	}

	// Try as single wallet
	var wallet WalletResponse
	if err := json.Unmarshal(rawResponse, &wallet); err != nil {
		return nil, fmt.Errorf("failed to decode wallet response: %w", err)
	}

	return &wallet, nil
}

//nolint:gosec // G404: math/rand acceptable for simulation test data generation
package service

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"math/rand"
	"time"

	"github.com/vnykmshr/nivo/services/simulation/internal/personas"
)

// UserStage represents the lifecycle stage of a simulated user
type UserStage string

const (
	StageNew          UserStage = "NEW"           // Ready to register
	StageRegistered   UserStage = "REGISTERED"    // Registered, needs KYC
	StageKYCSubmitted UserStage = "KYC_SUBMITTED" // KYC submitted, awaiting verification
	StageKYCVerified  UserStage = "KYC_VERIFIED"  // KYC verified, can transact
	StageActive       UserStage = "ACTIVE"        // Fully active user
)

// SimulatedUser represents a user created and managed by the simulation engine
type SimulatedUser struct {
	// Identity
	Email       string
	Password    string
	FullName    string
	PhoneNumber string

	// State
	UserID       string // Set after registration
	WalletID     string // Set after wallet creation
	SessionToken string // Set after login
	Persona      personas.PersonaType

	// Lifecycle tracking
	Stage      UserStage
	Balance    int64
	CreatedAt  time.Time
	LastActive time.Time
	LastLogin  time.Time
}

// UserLifecycleManager manages the lifecycle of simulated users
type UserLifecycleManager struct {
	gatewayClient *GatewayClient
	db            *sql.DB // Direct DB access for admin bypasses
	users         []*SimulatedUser
}

// NewUserLifecycleManager creates a new user lifecycle manager
func NewUserLifecycleManager(gatewayClient *GatewayClient, db *sql.DB) *UserLifecycleManager {
	return &UserLifecycleManager{
		gatewayClient: gatewayClient,
		db:            db,
		users:         make([]*SimulatedUser, 0),
	}
}

// GenerateNewUser creates a new simulated user with random persona
func (m *UserLifecycleManager) GenerateNewUser() *SimulatedUser {
	timestamp := time.Now().Unix()
	personaTypes := personas.AllPersonaTypes()
	persona := personaTypes[rand.Intn(len(personaTypes))]

	// Generate realistic Indian names
	firstNames := []string{"Amit", "Priya", "Rahul", "Sneha", "Vikram", "Anjali", "Arjun", "Kavya", "Rohan", "Diya"}
	lastNames := []string{"Sharma", "Patel", "Kumar", "Singh", "Reddy", "Nair", "Gupta", "Verma", "Iyer", "Mehta"}

	firstName := firstNames[rand.Intn(len(firstNames))]
	lastName := lastNames[rand.Intn(len(lastNames))]

	user := &SimulatedUser{
		Email:       fmt.Sprintf("%s.%s.%d@example.com", firstName, lastName, timestamp),
		Password:    generateRandomPassword(),
		FullName:    fmt.Sprintf("%s %s", firstName, lastName),
		PhoneNumber: generateIndianPhone(),
		Persona:     persona,
		Stage:       StageNew,
		CreatedAt:   time.Now(),
		LastActive:  time.Now(),
	}

	m.users = append(m.users, user)
	log.Printf("[simulation] 📝 Generated new user: %s (%s)", user.Email, user.Persona)
	return user
}

// RegisterUser registers a new user in the system
func (m *UserLifecycleManager) RegisterUser(ctx context.Context, user *SimulatedUser) error {
	if user.Stage != StageNew {
		return fmt.Errorf("user must be in NEW stage to register (current: %s)", user.Stage)
	}

	resp, err := m.gatewayClient.RegisterUser(ctx, user.Email, user.PhoneNumber, user.FullName, user.Password)
	if err != nil {
		return fmt.Errorf("failed to register user: %w", err)
	}

	user.UserID = resp.ID
	user.Stage = StageRegistered
	user.LastActive = time.Now()

	log.Printf("[simulation] ✅ User registered: %s (ID: %s)", user.Email, user.UserID)
	return nil
}

// LoginUser logs in a user and obtains a session token
func (m *UserLifecycleManager) LoginUser(ctx context.Context, user *SimulatedUser) error {
	if user.Stage == StageNew {
		return fmt.Errorf("user must be registered before logging in")
	}

	resp, err := m.gatewayClient.Login(ctx, user.Email, user.Password)
	if err != nil {
		return fmt.Errorf("failed to login user: %w", err)
	}

	user.SessionToken = resp.Token
	user.LastLogin = time.Now()
	user.LastActive = time.Now()

	// Fetch wallet ID if not already set
	if user.WalletID == "" && user.UserID != "" {
		wallet, walletErr := m.gatewayClient.GetUserWallet(ctx, resp.Token, user.UserID)
		if walletErr != nil {
			log.Printf("[simulation] Warning: Failed to fetch wallet for user %s: %v", user.Email, walletErr)
		} else {
			user.WalletID = wallet.ID
			user.Balance = wallet.Balance
			log.Printf("[simulation] 💼 Wallet fetched: %s (balance: ₹%.2f)", user.WalletID, float64(wallet.Balance)/100)
		}
	}

	log.Printf("[simulation] 🔐 User logged in: %s", user.Email)
	return nil
}

// SubmitKYC submits KYC information for a user
func (m *UserLifecycleManager) SubmitKYC(ctx context.Context, user *SimulatedUser) error {
	if user.Stage != StageRegistered {
		return fmt.Errorf("user must be in REGISTERED stage to submit KYC (current: %s)", user.Stage)
	}

	// Ensure user is logged in
	if user.SessionToken == "" {
		if err := m.LoginUser(ctx, user); err != nil {
			return fmt.Errorf("failed to login before KYC: %w", err)
		}
	}

	kycReq := generateKYCData(user.FullName)
	if err := m.gatewayClient.SubmitKYC(ctx, user.SessionToken, kycReq); err != nil {
		return fmt.Errorf("failed to submit KYC: %w", err)
	}

	user.Stage = StageKYCSubmitted
	user.LastActive = time.Now()

	log.Printf("[simulation] 📋 KYC submitted: %s", user.Email)
	return nil
}

// VerifyKYC uses local database bypass to verify a user's KYC (simulated users only)
func (m *UserLifecycleManager) VerifyKYC(ctx context.Context, user *SimulatedUser) error {
	if user.Stage != StageKYCSubmitted {
		return fmt.Errorf("user must be in KYC_SUBMITTED stage to verify (current: %s)", user.Stage)
	}

	// LOCAL BYPASS: Direct database update for simulated users
	// This bypasses the need for admin tokens and API calls
	if err := m.verifyKYCDirectly(ctx, user.UserID); err != nil {
		return fmt.Errorf("failed to verify KYC directly: %w", err)
	}

	user.Stage = StageKYCVerified
	user.LastActive = time.Now()

	log.Printf("[simulation] ✅ KYC verified (local bypass): %s", user.Email)
	return nil
}

// verifyKYCDirectly directly updates the KYC status in the database
// This is a local bypass for simulated users only - NOT for production use
func (m *UserLifecycleManager) verifyKYCDirectly(ctx context.Context, userID string) error {
	// Begin transaction for atomicity
	tx, err := m.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	// Update user_kyc table to set status to 'verified'
	kycQuery := `
		UPDATE user_kyc
		SET status = 'verified',
		    verified_at = NOW(),
		    updated_at = NOW()
		WHERE user_id = $1
		RETURNING id
	`

	var kycID string
	err = tx.QueryRowContext(ctx, kycQuery, userID).Scan(&kycID)
	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("no KYC record found for user %s", userID)
		}
		return fmt.Errorf("failed to update KYC status: %w", err)
	}

	// Update user status to 'active'
	userQuery := `
		UPDATE users
		SET status = 'active',
		    updated_at = NOW()
		WHERE id = $1
	`

	_, err = tx.ExecContext(ctx, userQuery, userID)
	if err != nil {
		return fmt.Errorf("failed to update user status: %w", err)
	}

	// Commit transaction
	if err = tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	log.Printf("[simulation] 🔧 Direct DB update: KYC verified for user %s (kyc_id: %s)", userID, kycID)
	return nil
}

// LogoutUser logs out a user
func (m *UserLifecycleManager) LogoutUser(ctx context.Context, user *SimulatedUser) error {
	if user.SessionToken == "" {
		return fmt.Errorf("user is not logged in")
	}

	if err := m.gatewayClient.Logout(ctx, user.SessionToken); err != nil {
		return fmt.Errorf("failed to logout user: %w", err)
	}

	user.SessionToken = ""
	user.LastActive = time.Now()

	log.Printf("[simulation] 🚪 User logged out: %s", user.Email)
	return nil
}

// GetUsers returns all simulated users
func (m *UserLifecycleManager) GetUsers() []*SimulatedUser {
	return m.users
}

// GetActiveUsers returns users who are KYC verified or active
func (m *UserLifecycleManager) GetActiveUsers() []*SimulatedUser {
	active := make([]*SimulatedUser, 0)
	for _, u := range m.users {
		if u.Stage == StageKYCVerified || u.Stage == StageActive {
			active = append(active, u)
		}
	}
	return active
}

// Helper functions

func generateRandomPassword() string {
	// Generate a simple but valid password
	return fmt.Sprintf("Pass%d!@#", rand.Intn(100000)+10000)
}

func generateIndianPhone() string {
	// Generate a valid Indian mobile number (starts with 6-9)
	firstDigit := rand.Intn(4) + 6 // 6, 7, 8, or 9
	remaining := rand.Intn(900000000) + 100000000
	return fmt.Sprintf("%d%09d", firstDigit, remaining)
}

func generateKYCData(_ string) KYCSubmitRequest {
	// Generate realistic Indian addresses
	streets := []string{"MG Road", "Brigade Road", "Residency Road", "Church Street", "Indiranagar"}
	cities := []string{"Bangalore", "Mumbai", "Delhi", "Chennai", "Hyderabad", "Pune", "Kolkata"}
	states := []string{"Karnataka", "Maharashtra", "Delhi", "Tamil Nadu", "Telangana", "Maharashtra", "West Bengal"}

	cityIdx := rand.Intn(len(cities))

	// Generate valid DOB (1980-2000 range, valid month 1-12, valid day 1-28)
	year := 1980 + rand.Intn(21) // 1980-2000
	month := rand.Intn(12) + 1   // 1-12
	day := rand.Intn(28) + 1     // 1-28 (safe for all months)

	return KYCSubmitRequest{
		PAN:         generatePAN(),
		Aadhaar:     generateAadhar(),
		DateOfBirth: fmt.Sprintf("%04d-%02d-%02d", year, month, day),
		Address: KYCAddressRequest{
			Street:  fmt.Sprintf("%d, %s", rand.Intn(500)+1, streets[rand.Intn(len(streets))]),
			City:    cities[cityIdx],
			State:   states[cityIdx],
			PIN:     fmt.Sprintf("%06d", rand.Intn(900000)+100000),
			Country: "IN",
		},
	}
}

func generatePAN() string {
	// PAN format: AAAAA9999A
	letters := "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	pan := ""
	for i := 0; i < 5; i++ {
		pan += string(letters[rand.Intn(len(letters))])
	}
	pan += fmt.Sprintf("%04d", rand.Intn(10000))
	pan += string(letters[rand.Intn(len(letters))])
	return pan
}

func generateAadhar() string {
	// Aadhar format: 12 digits
	return fmt.Sprintf("%012d", rand.Intn(900000000000)+100000000000)
}

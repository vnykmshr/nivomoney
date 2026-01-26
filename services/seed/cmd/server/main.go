package main

import (
	"context"
	"crypto/rand"
	_ "embed"
	"encoding/base64"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
	"gopkg.in/yaml.v3"

	"github.com/vnykmshr/nivo/shared/config"
	"github.com/vnykmshr/nivo/shared/database"
)

//go:embed data/users.yaml
var usersData []byte

const serviceName = "seed"

// SeedData represents the structure of seed data files
type SeedData struct {
	Users []SeedUser `yaml:"users"`
}

// SeedUser represents a user in the seed data
type SeedUser struct {
	ID             string  `yaml:"id"`
	FullName       string  `yaml:"full_name"`
	Email          string  `yaml:"email"`
	Password       string  `yaml:"password"`
	Phone          string  `yaml:"phone"`
	PAN            string  `yaml:"pan"`
	Aadhaar        string  `yaml:"aadhaar"`
	DateOfBirth    string  `yaml:"date_of_birth"`
	Address        Address `yaml:"address"`
	InitialBalance int64   `yaml:"initial_balance"` // In paise
}

// Address represents user address
type Address struct {
	Street  string `yaml:"street" json:"street"`
	City    string `yaml:"city" json:"city"`
	State   string `yaml:"state" json:"state"`
	PIN     string `yaml:"pin" json:"pin"`
	Country string `yaml:"country" json:"country"`
}

func main() {
	// Parse command line flags
	cleanFlag := flag.Bool("clean", false, "Clean database before seeding")
	flag.Parse()

	log.Printf("[%s] ========================================", serviceName)
	log.Printf("[%s]   Nivo Money - Database Seed Script", serviceName)
	log.Printf("[%s] ========================================", serviceName)
	log.Printf("[%s] Clean mode: %v", serviceName, *cleanFlag)

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("[%s] Failed to load configuration: %v", serviceName, err)
	}

	// Connect to database
	db, err := database.NewFromURL(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("[%s] Failed to connect to database: %v", serviceName, err)
	}
	defer func() { _ = db.Close() }()

	log.Printf("[%s] Connected to database successfully", serviceName)

	// Clean database if requested
	if *cleanFlag {
		if err := cleanDatabase(db); err != nil {
			log.Fatalf("[%s] Failed to clean database: %v", serviceName, err)
		}
		log.Printf("[%s] Database cleaned successfully", serviceName)
	}

	// Parse seed data
	var seedData SeedData
	if err := yaml.Unmarshal(usersData, &seedData); err != nil {
		log.Fatalf("[%s] Failed to parse seed data: %v", serviceName, err)
	}

	log.Printf("[%s] Loaded %d users from seed data", serviceName, len(seedData.Users))

	// Generate secure admin password (override static YAML value)
	adminPassword := generateSecurePassword(16)
	for i := range seedData.Users {
		if seedData.Users[i].Email == "admin@nivo.local" || seedData.Users[i].Email == "admin@vnykmshr.com" {
			seedData.Users[i].Password = adminPassword
			log.Printf("[%s] Generated secure admin password (not using static YAML value)", serviceName)
		}
	}

	log.Printf("[%s] ", serviceName)

	ctx := context.Background()

	// Setup role permissions (ensure 'user' role has required permissions)
	if err := setupRolePermissions(ctx, db); err != nil {
		log.Fatalf("[%s] Failed to setup role permissions: %v", serviceName, err)
	}

	// Seed users with complete setup
	if err := seedCompleteUsers(ctx, db, seedData.Users); err != nil {
		log.Fatalf("[%s] Failed to seed users: %v", serviceName, err)
	}

	// Ensure all wallets have limits (for existing wallets that may not have limits)
	if err := ensureWalletLimits(ctx, db); err != nil {
		log.Printf("[%s] Warning: Failed to ensure wallet limits: %v", serviceName, err)
	}

	// Seed demo data (beneficiaries, virtual cards)
	if err := seedDemoData(ctx, db); err != nil {
		log.Printf("[%s] Warning: Failed to seed demo data: %v", serviceName, err)
	}

	// Write credentials to .secrets/credentials.txt
	if err := writeCredentialsFile(seedData.Users, adminPassword); err != nil {
		log.Printf("[%s] Warning: Failed to write credentials file: %v", serviceName, err)
	}

	log.Printf("[%s] ", serviceName)
	log.Printf("[%s] ========================================", serviceName)
	log.Printf("[%s] Seed completed successfully!", serviceName)
	log.Printf("[%s] Created/verified %d ready-to-use accounts", serviceName, len(seedData.Users))
	log.Printf("[%s] ========================================", serviceName)
	log.Printf("[%s] ", serviceName)
	log.Printf("[%s] ┌─────────────────────────────────────────────────────┐", serviceName)
	log.Printf("[%s] │            ADMIN CREDENTIALS (Generated)            │", serviceName)
	log.Printf("[%s] ├─────────────────────────────────────────────────────┤", serviceName)
	log.Printf("[%s] │  Email:    admin@nivo.local                         │", serviceName)
	log.Printf("[%s] │  Password: %-40s │", serviceName, adminPassword)
	log.Printf("[%s] ├─────────────────────────────────────────────────────┤", serviceName)
	log.Printf("[%s] │  Credentials saved to: .secrets/credentials.txt    │", serviceName)
	log.Printf("[%s] └─────────────────────────────────────────────────────┘", serviceName)
	log.Printf("[%s] ", serviceName)
	log.Printf("[%s] Demo user credentials are in README.md (public)", serviceName)
}

// seedCompleteUsers creates complete user accounts with KYC, wallets, and initial balance
func seedCompleteUsers(ctx context.Context, db *database.DB, users []SeedUser) error {
	log.Printf("[%s] ========== Seeding Complete User Accounts ==========", serviceName)

	for i, seedUser := range users {
		log.Printf("[%s] ", serviceName)
		log.Printf("[%s] [%d/%d] Processing: %s (%s)", serviceName, i+1, len(users), seedUser.FullName, seedUser.Email)

		// Step 1: Create or get user
		userID, err := createUser(ctx, db, seedUser)
		if err != nil {
			log.Printf("[%s]   ERROR: Failed to create user: %v", serviceName, err)
			continue
		}

		// Step 1b: Create User-Admin account (skip for admin users)
		isAdmin := seedUser.Email == "admin@vnykmshr.com" || seedUser.Email == "admin@nivo.local"
		var userAdminID string
		if !isAdmin {
			userAdminID, err = createUserAdmin(ctx, db, userID, seedUser)
			if err != nil {
				log.Printf("[%s]   ERROR: Failed to create User-Admin: %v", serviceName, err)
				// Continue without User-Admin - not critical for seeding
			}

			// Step 1c: Create User-Admin pairing
			if userAdminID != "" {
				if err := createUserAdminPairing(ctx, db, userID, userAdminID); err != nil {
					log.Printf("[%s]   ERROR: Failed to create User-Admin pairing: %v", serviceName, err)
				}
			}
		}

		// Step 2: Create or update KYC (verified status)
		if err := createKYC(ctx, db, userID, seedUser); err != nil {
			log.Printf("[%s]   ERROR: Failed to create KYC: %v", serviceName, err)
			continue
		}

		// Step 3: Update user status to active
		if err := activateUser(ctx, db, userID); err != nil {
			log.Printf("[%s]   ERROR: Failed to activate user: %v", serviceName, err)
			continue
		}

		// Step 3b: Activate User-Admin account too
		if userAdminID != "" {
			if err := activateUser(ctx, db, userAdminID); err != nil {
				log.Printf("[%s]   ERROR: Failed to activate User-Admin: %v", serviceName, err)
			}
		}

		// Step 4: Assign roles
		// - Regular users get 'user' role
		// - Admin users get 'user', 'admin', and 'super_admin' roles
		if isAdmin {
			for _, roleName := range []string{"user", "admin", "super_admin"} {
				if err := assignUserRole(ctx, db, userID, roleName); err != nil {
					log.Printf("[%s]   ERROR: Failed to assign %s role: %v", serviceName, roleName, err)
				}
			}
		} else {
			if err := assignUserRole(ctx, db, userID, "user"); err != nil {
				log.Printf("[%s]   ERROR: Failed to assign role: %v", serviceName, err)
				continue
			}
		}

		// Step 4b: Assign 'user_admin' role to User-Admin account
		if userAdminID != "" {
			if err := assignUserRole(ctx, db, userAdminID, "user_admin"); err != nil {
				log.Printf("[%s]   ERROR: Failed to assign user_admin role: %v", serviceName, err)
			}
		}

		// Step 5: Create ledger account for wallet
		ledgerAccountID, err := createLedgerAccount(ctx, db, userID, seedUser)
		if err != nil {
			log.Printf("[%s]   ERROR: Failed to create ledger account: %v", serviceName, err)
			continue
		}

		// Step 6: Create wallet
		walletID, err := createWallet(ctx, db, userID, ledgerAccountID)
		if err != nil {
			log.Printf("[%s]   ERROR: Failed to create wallet: %v", serviceName, err)
			continue
		}

		// Step 7: Add initial balance if specified
		if seedUser.InitialBalance > 0 {
			if err := addInitialBalance(ctx, db, userID, walletID, ledgerAccountID, seedUser.InitialBalance); err != nil {
				log.Printf("[%s]   ERROR: Failed to add initial balance: %v", serviceName, err)
				continue
			}
		}

		adminInfo := ""
		if userAdminID != "" {
			adminInfo = ", UserAdmin=paired"
		}
		log.Printf("[%s]   ✓ Complete account ready: UserID=%s, WalletID=%s, Balance=₹%.2f%s",
			serviceName, userID, walletID, float64(seedUser.InitialBalance)/100.0, adminInfo)
	}

	log.Printf("[%s] ", serviceName)
	return nil
}

// createUser creates a user or returns existing user ID
func createUser(ctx context.Context, db *database.DB, user SeedUser) (string, error) {
	// Check if user exists (by email AND account_type since same email can have multiple account types)
	var existingID string
	err := db.QueryRowContext(ctx,
		"SELECT id FROM users WHERE email = $1 AND account_type = 'user'",
		user.Email).Scan(&existingID)
	if err == nil {
		log.Printf("[%s]   → User already exists: %s", serviceName, existingID)
		return existingID, nil
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}

	// Insert user with account_type = 'user'
	var userID string
	query := `
		INSERT INTO users (email, phone, full_name, password_hash, status, account_type)
		VALUES ($1, $2, $3, $4, 'pending', 'user')
		RETURNING id
	`
	err = db.QueryRowContext(ctx, query, user.Email, user.Phone, user.FullName, string(hashedPassword)).Scan(&userID)
	if err != nil {
		return "", err
	}

	log.Printf("[%s]   → User created: %s", serviceName, userID)
	return userID, nil
}

// createUserAdmin creates a User-Admin account paired with the regular user
// Uses the same email as the regular user - uniqueness is (email, account_type) composite
func createUserAdmin(ctx context.Context, db *database.DB, userID string, user SeedUser) (string, error) {
	// Check if User-Admin exists (by email AND account_type)
	var existingID string
	err := db.QueryRowContext(ctx,
		"SELECT id FROM users WHERE email = $1 AND account_type = 'user_admin'",
		user.Email).Scan(&existingID)
	if err == nil {
		log.Printf("[%s]   → User-Admin already exists: %s", serviceName, existingID)
		return existingID, nil
	}

	// Hash password (same as regular user)
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}

	// Insert User-Admin with same email and account_type = 'user_admin', NULL phone
	var userAdminID string
	query := `
		INSERT INTO users (email, phone, full_name, password_hash, status, account_type)
		VALUES ($1, NULL, $2, $3, 'pending', 'user_admin')
		RETURNING id
	`
	err = db.QueryRowContext(ctx, query, user.Email, user.FullName+" (Admin)", string(hashedPassword)).Scan(&userAdminID)
	if err != nil {
		return "", err
	}

	log.Printf("[%s]   → User-Admin created: %s (same email, account_type=user_admin)", serviceName, userAdminID)
	return userAdminID, nil
}

// createUserAdminPairing creates the pairing between user and User-Admin
func createUserAdminPairing(ctx context.Context, db *database.DB, userID, userAdminID string) error {
	// Check if pairing exists
	var existingID string
	err := db.QueryRowContext(ctx,
		"SELECT id FROM user_admin_pairs WHERE user_id = $1 AND admin_user_id = $2",
		userID, userAdminID).Scan(&existingID)
	if err == nil {
		log.Printf("[%s]   → User-Admin pairing already exists", serviceName)
		return nil
	}

	// Insert pairing
	query := `
		INSERT INTO user_admin_pairs (user_id, admin_user_id)
		VALUES ($1, $2)
	`
	_, err = db.ExecContext(ctx, query, userID, userAdminID)
	if err != nil {
		return err
	}

	log.Printf("[%s]   → User-Admin pairing created", serviceName)
	return nil
}

// createKYC creates or updates KYC information with verified status
func createKYC(ctx context.Context, db *database.DB, userID string, user SeedUser) error {
	addressJSON, err := json.Marshal(user.Address)
	if err != nil {
		return err
	}

	// Check if KYC exists
	var existingStatus string
	existingErr := db.QueryRowContext(ctx, "SELECT status FROM user_kyc WHERE user_id = $1", userID).Scan(&existingStatus)

	if existingErr == nil {
		// Update existing KYC to verified
		query := `
			UPDATE user_kyc
			SET status = 'verified',
			    pan = $2,
			    aadhaar = $3,
			    date_of_birth = $4,
			    address = $5,
			    verified_at = NOW(),
			    updated_at = NOW()
			WHERE user_id = $1
		`
		_, err = db.ExecContext(ctx, query, userID, user.PAN, user.Aadhaar, user.DateOfBirth, addressJSON)
		if err != nil {
			return err
		}
		log.Printf("[%s]   → KYC updated to verified", serviceName)
		return nil
	}

	// Insert new KYC with verified status
	query := `
		INSERT INTO user_kyc (user_id, status, pan, aadhaar, date_of_birth, address, verified_at)
		VALUES ($1, 'verified', $2, $3, $4, $5, NOW())
	`
	_, err = db.ExecContext(ctx, query, userID, user.PAN, user.Aadhaar, user.DateOfBirth, addressJSON)
	if err != nil {
		return err
	}

	log.Printf("[%s]   → KYC created and verified", serviceName)
	return nil
}

// activateUser updates user status to active
func activateUser(ctx context.Context, db *database.DB, userID string) error {
	query := "UPDATE users SET status = 'active', updated_at = NOW() WHERE id = $1"
	_, err := db.ExecContext(ctx, query, userID)
	if err != nil {
		return err
	}
	log.Printf("[%s]   → User activated", serviceName)
	return nil
}

// createLedgerAccount creates a ledger account for the user's wallet
func createLedgerAccount(ctx context.Context, db *database.DB, userID string, user SeedUser) (string, error) {
	// Check if ledger account already exists for this user
	var existingID string
	checkQuery := `
		SELECT id FROM accounts
		WHERE code LIKE $1 AND status = 'active'
		LIMIT 1
	`
	codePattern := "WALLET-" + userID[:8] + "%"
	err := db.QueryRowContext(ctx, checkQuery, codePattern).Scan(&existingID)
	if err == nil {
		log.Printf("[%s]   → Ledger account already exists: %s", serviceName, existingID)
		return existingID, nil
	}

	// Create new ledger account under Customer Deposits (liability)
	var accountID string
	query := `
		INSERT INTO accounts (code, name, type, currency, status, metadata)
		VALUES ($1, $2, 'liability', 'INR', 'active', $3)
		RETURNING id
	`
	code := "WALLET-" + userID
	name := user.FullName + " - Wallet"
	metadata := map[string]interface{}{
		"user_id":  userID,
		"purpose":  "customer_wallet",
		"category": "customer_deposits",
	}
	metadataJSON, _ := json.Marshal(metadata)

	err = db.QueryRowContext(ctx, query, code, name, metadataJSON).Scan(&accountID)
	if err != nil {
		return "", err
	}

	log.Printf("[%s]   → Ledger account created: %s", serviceName, accountID)
	return accountID, nil
}

// createWallet creates a wallet for the user
func createWallet(ctx context.Context, db *database.DB, userID, ledgerAccountID string) (string, error) {
	// Check if wallet already exists
	var existingID string
	err := db.QueryRowContext(ctx,
		"SELECT id FROM wallets WHERE user_id = $1 AND type = 'default' AND status != 'closed'",
		userID).Scan(&existingID)
	if err == nil {
		log.Printf("[%s]   → Wallet already exists: %s", serviceName, existingID)
		return existingID, nil
	}

	// Create new wallet
	var walletID string
	query := `
		INSERT INTO wallets (user_id, type, currency, balance, available_balance, status, ledger_account_id)
		VALUES ($1, 'default', 'INR', 0, 0, 'active', $2)
		RETURNING id
	`
	err = db.QueryRowContext(ctx, query, userID, ledgerAccountID).Scan(&walletID)
	if err != nil {
		return "", err
	}

	// Create default wallet limits (₹10,000/day = 1000000 paise, ₹100,000/month = 10000000 paise)
	limitsQuery := `
		INSERT INTO wallet_limits (wallet_id, daily_limit, monthly_limit)
		VALUES ($1, 1000000, 10000000)
	`
	_, err = db.ExecContext(ctx, limitsQuery, walletID)
	if err != nil {
		log.Printf("[%s]   Warning: Failed to create wallet limits: %v", serviceName, err)
		// Continue - limits can be created later
	}

	log.Printf("[%s]   → Wallet created: %s (with default limits)", serviceName, walletID)
	return walletID, nil
}

// addInitialBalance adds initial balance to wallet using proper double-entry bookkeeping
func addInitialBalance(ctx context.Context, db *database.DB, userID, walletID, ledgerAccountID string, amount int64) error {
	// Start transaction
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	// Create journal entry with unique entry number (includes user ID for uniqueness)
	var journalEntryID string
	entryNumber := fmt.Sprintf("SEED-%s-%s", time.Now().Format("20060102-150405"), userID[:8])
	journalQuery := `
		INSERT INTO journal_entries (entry_number, type, status, description, reference_type, reference_id, posted_at, posted_by)
		VALUES ($1, 'opening', 'posted', $2, 'seed', $3, NOW(), $4::uuid)
		RETURNING id
	`
	description := "Initial seed deposit for " + userID[:8]
	err = tx.QueryRowContext(ctx, journalQuery, entryNumber, description, userID, userID).Scan(&journalEntryID)
	if err != nil {
		log.Printf("[%s]   DEBUG: Journal entry creation failed. Entry: %s, Desc: %s, UserID: %s, Error: %v",
			serviceName, entryNumber, description, userID, err)
		return fmt.Errorf("journal entry creation failed: %w", err)
	}
	log.Printf("[%s]   DEBUG: Journal entry created: %s", serviceName, journalEntryID)

	// Get or create cash account (debit source)
	var cashAccountID string
	err = tx.QueryRowContext(ctx, "SELECT id FROM accounts WHERE code = '1000'").Scan(&cashAccountID)
	if err != nil {
		return err
	}

	// Create ledger lines (double-entry)
	// Debit: Cash account (asset decreases - but for seed we're creating money)
	// Credit: Customer deposit (liability increases)
	ledgerLineQuery := `
		INSERT INTO ledger_lines (entry_id, account_id, debit_amount, credit_amount, description)
		VALUES ($1, $2, $3, $4, $5)
	`

	// Debit cash (we're funding from our cash reserves)
	_, err = tx.ExecContext(ctx, ledgerLineQuery, journalEntryID, cashAccountID, amount, 0, "Seed deposit - debit cash")
	if err != nil {
		return err
	}

	// Credit customer wallet account (liability increases)
	_, err = tx.ExecContext(ctx, ledgerLineQuery, journalEntryID, ledgerAccountID, 0, amount, "Seed deposit - credit wallet")
	if err != nil {
		return err
	}

	// Update account balances
	// Cash account: debit increases asset balance
	_, err = tx.ExecContext(ctx, `
		UPDATE accounts
		SET balance = balance + $1,
		    debit_total = debit_total + $1,
		    updated_at = NOW()
		WHERE id = $2
	`, amount, cashAccountID)
	if err != nil {
		return err
	}

	// Wallet account: credit increases liability balance
	_, err = tx.ExecContext(ctx, `
		UPDATE accounts
		SET balance = balance + $1,
		    credit_total = credit_total + $1,
		    updated_at = NOW()
		WHERE id = $2
	`, amount, ledgerAccountID)
	if err != nil {
		return err
	}

	// Update wallet balance
	_, err = tx.ExecContext(ctx, `
		UPDATE wallets
		SET balance = balance + $1,
		    available_balance = available_balance + $1,
		    updated_at = NOW()
		WHERE id = $2
	`, amount, walletID)
	if err != nil {
		return err
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return err
	}

	log.Printf("[%s]   → Initial balance added: ₹%.2f", serviceName, float64(amount)/100.0)
	return nil
}

// setupRolePermissions ensures the 'user' and 'admin' roles have required permissions
func setupRolePermissions(ctx context.Context, db *database.DB) error {
	log.Printf("[%s] ========== Setting up Role Permissions ==========", serviceName)

	// User role permissions (basic user permissions)
	userPermissions := []string{
		"identity:auth:login",
		"identity:auth:logout",
		"identity:auth:refresh",
		"identity:profile:read",
		"identity:profile:update",
		"identity:kyc:submit",
		"identity:kyc:read",
		"wallet:wallet:create",
		"wallet:wallet:read",
		"wallet:wallet:list",
		"wallet:beneficiary:manage",
		"transaction:deposit:create",
		"transaction:transfer:create",
		"transaction:transaction:list",
		"transaction:transaction:read",
	}

	// Admin role permissions (everything)
	adminPermissions := []string{
		"identity:auth:login",
		"identity:auth:logout",
		"identity:auth:refresh",
		"identity:profile:read",
		"identity:profile:update",
		"identity:profile:delete",
		"identity:users:read",
		"identity:users:create",
		"identity:users:update",
		"identity:users:delete",
		"identity:kyc:submit",
		"identity:kyc:read",
		"identity:kyc:verify",
		"identity:kyc:reject",
		"identity:kyc:list",
		"identity:user:suspend",
		"identity:user:unsuspend",
		"wallet:wallet:create",
		"wallet:wallet:read",
		"wallet:wallet:update",
		"wallet:wallet:delete",
		"wallet:wallet:list",
		"wallet:wallet:freeze",
		"wallet:wallet:unfreeze",
		"wallet:beneficiary:manage",
		"transaction:deposit:create",
		"transaction:transfer:create",
		"transaction:withdrawal:create",
		"transaction:transaction:create",
		"transaction:transaction:list",
		"transaction:transaction:read",
		"transaction:transaction:reverse",
	}

	// Assign permissions to 'user' role
	userRoleID := "00000000-0000-0000-0000-000000000001"
	if err := assignPermissionsToRole(ctx, db, userRoleID, "user", userPermissions); err != nil {
		return err
	}

	// Assign permissions to 'admin' role
	adminRoleID := "00000000-0000-0000-0000-000000000005"
	if err := assignPermissionsToRole(ctx, db, adminRoleID, "admin", adminPermissions); err != nil {
		return err
	}

	log.Printf("[%s] ", serviceName)
	return nil
}

// assignPermissionsToRole assigns a list of permissions to a role
func assignPermissionsToRole(ctx context.Context, db *database.DB, roleID, roleName string, permissions []string) error {
	for _, permName := range permissions {
		// Get permission ID by name
		var permID string
		err := db.QueryRowContext(ctx, "SELECT id FROM permissions WHERE name = $1", permName).Scan(&permID)
		if err != nil {
			log.Printf("[%s]   WARNING: Permission '%s' not found, skipping", serviceName, permName)
			continue
		}

		// Check if already assigned
		var existingID string
		existErr := db.QueryRowContext(ctx,
			"SELECT role_id FROM role_permissions WHERE role_id = $1 AND permission_id = $2",
			roleID, permID).Scan(&existingID)
		if existErr == nil {
			continue // Already assigned
		}

		// Assign permission to role
		_, insertErr := db.ExecContext(ctx,
			"INSERT INTO role_permissions (role_id, permission_id) VALUES ($1, $2)",
			roleID, permID)
		if insertErr != nil {
			log.Printf("[%s]   WARNING: Failed to assign %s to %s: %v", serviceName, permName, roleName, insertErr)
			continue
		}
	}

	// Count assigned permissions
	var count int
	_ = db.QueryRowContext(ctx, "SELECT COUNT(*) FROM role_permissions WHERE role_id = $1", roleID).Scan(&count)
	log.Printf("[%s]   → Role '%s' now has %d permissions", serviceName, roleName, count)

	return nil
}

// assignUserRole assigns a role to a user
func assignUserRole(ctx context.Context, db *database.DB, userID, roleName string) error {
	// Get role ID by name
	var roleID string
	err := db.QueryRowContext(ctx, "SELECT id FROM roles WHERE name = $1", roleName).Scan(&roleID)
	if err != nil {
		return fmt.Errorf("role '%s' not found: %w", roleName, err)
	}

	// Check if already assigned
	var existingID string
	existErr := db.QueryRowContext(ctx,
		"SELECT user_id FROM user_roles WHERE user_id = $1 AND role_id = $2",
		userID, roleID).Scan(&existingID)
	if existErr == nil {
		log.Printf("[%s]   → Role '%s' already assigned", serviceName, roleName)
		return nil
	}

	// Assign role to user
	_, insertErr := db.ExecContext(ctx,
		"INSERT INTO user_roles (user_id, role_id) VALUES ($1, $2)",
		userID, roleID)
	if insertErr != nil {
		return fmt.Errorf("failed to assign role: %w", insertErr)
	}

	log.Printf("[%s]   → Role '%s' assigned", serviceName, roleName)
	return nil
}

// cleanDatabase truncates all tables except system users and chart of accounts
func cleanDatabase(db *database.DB) error {
	log.Printf("[%s] Cleaning database tables...", serviceName)

	// SQL to clean tables while preserving chart of accounts
	cleanSQL := `
		-- Truncate transaction-related tables
		TRUNCATE TABLE
			beneficiaries,
			processed_transfers,
			transactions,
			wallet_limits,
			wallets,
			user_kyc,
			user_roles,
			role_permissions,
			sessions,
			notifications,
			risk_events
		CASCADE;

		-- Clean ledger tables but preserve chart of accounts
		DELETE FROM ledger_lines;
		DELETE FROM journal_entries;

		-- Delete user-specific ledger accounts only (preserve chart of accounts 1000-5999)
		DELETE FROM accounts WHERE code LIKE 'WALLET-%';

		-- Reset chart of accounts balances to zero
		UPDATE accounts
		SET balance = 0,
		    debit_total = 0,
		    credit_total = 0,
		    updated_at = NOW()
		WHERE code NOT LIKE 'WALLET-%';

		-- Delete test users (preserve system users with @vnykmshr.com)
		DELETE FROM users WHERE email NOT LIKE '%@vnykmshr.com';
	`

	_, err := db.Exec(cleanSQL)
	return err
}

// ensureWalletLimits creates default limits for any wallets that don't have them
func ensureWalletLimits(ctx context.Context, db *database.DB) error {
	log.Printf("[%s] Ensuring wallet limits exist for all wallets...", serviceName)

	// Find wallets without limits and create default limits
	query := `
		INSERT INTO wallet_limits (wallet_id, daily_limit, monthly_limit)
		SELECT w.id, 1000000, 10000000
		FROM wallets w
		LEFT JOIN wallet_limits wl ON w.id = wl.wallet_id
		WHERE wl.id IS NULL
	`

	result, err := db.ExecContext(ctx, query)
	if err != nil {
		return err
	}

	count, _ := result.RowsAffected()
	if count > 0 {
		log.Printf("[%s]   → Created default limits for %d wallets", serviceName, count)
	} else {
		log.Printf("[%s]   → All wallets already have limits", serviceName)
	}

	return nil
}

// seedDemoData creates demo beneficiaries and virtual cards for existing users
func seedDemoData(ctx context.Context, db *database.DB) error {
	log.Printf("[%s] ========== Seeding Demo Data ==========", serviceName)

	// Get all non-admin users
	rows, err := db.QueryContext(ctx, `
		SELECT u.id, u.full_name, w.id as wallet_id
		FROM users u
		JOIN wallets w ON w.user_id = u.id
		WHERE u.account_type = 'user' AND u.email NOT LIKE '%admin%'
		AND u.status = 'active'
	`)
	if err != nil {
		return fmt.Errorf("failed to query users: %w", err)
	}
	defer func() { _ = rows.Close() }()

	type userWallet struct {
		UserID   string
		FullName string
		WalletID string
	}

	var users []userWallet
	for rows.Next() {
		var uw userWallet
		if err := rows.Scan(&uw.UserID, &uw.FullName, &uw.WalletID); err != nil {
			continue
		}
		users = append(users, uw)
	}

	log.Printf("[%s] Found %d users for demo data seeding", serviceName, len(users))

	// Seed beneficiaries for each user
	for _, user := range users {
		if err := seedBeneficiariesForUser(ctx, db, user.UserID); err != nil {
			log.Printf("[%s]   Warning: Failed to seed beneficiaries for %s: %v", serviceName, user.FullName, err)
		}

		if err := seedVirtualCardForUser(ctx, db, user.WalletID, user.UserID, user.FullName); err != nil {
			log.Printf("[%s]   Warning: Failed to seed virtual card for %s: %v", serviceName, user.FullName, err)
		}
	}

	log.Printf("[%s] Demo data seeding complete", serviceName)
	return nil
}

// seedBeneficiariesForUser creates sample beneficiaries for a user
func seedBeneficiariesForUser(ctx context.Context, db *database.DB, userID string) error {
	// Sample beneficiaries
	beneficiaries := []struct {
		Nickname    string
		AccountID   string
		AccountType string
		IsFavorite  bool
	}{
		{
			Nickname:    "My Savings",
			AccountID:   "savings_" + userID[:8],
			AccountType: "nivo_wallet",
			IsFavorite:  true,
		},
		{
			Nickname:    "Rent Payment",
			AccountID:   "landlord_rent@hdfc",
			AccountType: "upi",
			IsFavorite:  false,
		},
		{
			Nickname:    "Electricity Bill",
			AccountID:   "electric.bill@paytm",
			AccountType: "upi",
			IsFavorite:  false,
		},
	}

	for _, b := range beneficiaries {
		// Check if beneficiary already exists
		var existingID string
		err := db.QueryRowContext(ctx,
			"SELECT id FROM beneficiaries WHERE user_id = $1 AND nickname = $2",
			userID, b.Nickname).Scan(&existingID)
		if err == nil {
			continue // Already exists
		}

		// Insert beneficiary
		query := `
			INSERT INTO beneficiaries (user_id, nickname, account_identifier, account_type, is_favorite, is_verified)
			VALUES ($1, $2, $3, $4, $5, true)
		`
		_, insertErr := db.ExecContext(ctx, query, userID, b.Nickname, b.AccountID, b.AccountType, b.IsFavorite)
		if insertErr != nil {
			log.Printf("[%s]     Warning: Failed to create beneficiary %s: %v", serviceName, b.Nickname, insertErr)
			continue
		}
	}

	log.Printf("[%s]   → Beneficiaries created for user %s", serviceName, userID[:8])
	return nil
}

// seedVirtualCardForUser creates a virtual card for a user's wallet
func seedVirtualCardForUser(ctx context.Context, db *database.DB, walletID, userID, fullName string) error {
	// Check if user already has a virtual card
	var existingID string
	err := db.QueryRowContext(ctx,
		"SELECT id FROM virtual_cards WHERE wallet_id = $1",
		walletID).Scan(&existingID)
	if err == nil {
		log.Printf("[%s]   → Virtual card already exists for wallet %s", serviceName, walletID[:8])
		return nil
	}

	// Generate card details
	cardNumber := generateCardNumber()
	expiryMonth := 12
	expiryYear := time.Now().Year() + 3 // 3 years validity
	cvv := generateCVV()

	// Hash CVV for storage
	cvvHash, err := bcrypt.GenerateFromPassword([]byte(cvv), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash CVV: %w", err)
	}

	// Insert virtual card
	query := `
		INSERT INTO virtual_cards (
			wallet_id, user_id, card_number, card_holder_name,
			expiry_month, expiry_year, cvv, card_type, status,
			daily_limit, monthly_limit, per_transaction_limit
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, 'virtual', 'active',
			5000000, 50000000, 2500000
		)
	`
	_, insertErr := db.ExecContext(ctx, query,
		walletID, userID, cardNumber, fullName,
		expiryMonth, expiryYear, string(cvvHash))
	if insertErr != nil {
		return fmt.Errorf("failed to create virtual card: %w", insertErr)
	}

	log.Printf("[%s]   → Virtual card created for %s (****%s)", serviceName, fullName, cardNumber[12:])
	return nil
}

// generateCardNumber generates a valid-looking 16-digit card number
func generateCardNumber() string {
	// Start with Nivo's test BIN (4000 for Visa-like test cards)
	prefix := "4000"
	// Generate 11 random digits
	middle := fmt.Sprintf("%011d", time.Now().UnixNano()%100000000000)
	partial := prefix + middle

	// Calculate Luhn check digit
	checkDigit := calculateLuhnCheckDigit(partial)
	return partial + fmt.Sprintf("%d", checkDigit)
}

// calculateLuhnCheckDigit calculates the Luhn check digit
func calculateLuhnCheckDigit(number string) int {
	sum := 0
	for i := len(number) - 1; i >= 0; i-- {
		digit := int(number[i] - '0')
		if (len(number)-i)%2 == 1 {
			digit *= 2
			if digit > 9 {
				digit -= 9
			}
		}
		sum += digit
	}
	return (10 - (sum % 10)) % 10
}

// generateCVV generates a 3-digit CVV
func generateCVV() string {
	return fmt.Sprintf("%03d", time.Now().UnixNano()%1000)
}

// generateSecurePassword generates a cryptographically secure random password
func generateSecurePassword(length int) string {
	// Generate random bytes
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		// Fallback to time-based if crypto/rand fails (should never happen)
		return fmt.Sprintf("Admin%d!", time.Now().UnixNano()%100000000)
	}
	// Use URL-safe base64 encoding and trim to desired length
	password := base64.URLEncoding.EncodeToString(bytes)
	// Ensure we have at least the requested length
	if len(password) > length {
		password = password[:length]
	}
	return password
}

// SeededCredentials holds the credentials for all seeded users
type SeededCredentials struct {
	GeneratedAt   string                 `json:"generated_at"`
	AdminPassword string                 `json:"admin_password"`
	Users         []SeededUserCredential `json:"users"`
}

// SeededUserCredential holds a single user's credentials
type SeededUserCredential struct {
	Email       string `json:"email"`
	Password    string `json:"password"`
	Role        string `json:"role"`
	FullName    string `json:"full_name"`
	IsGenerated bool   `json:"is_generated"`
}

// writeCredentialsFile writes all seed credentials to .secrets/credentials.txt
func writeCredentialsFile(users []SeedUser, adminPassword string) error {
	// Create .secrets directory if it doesn't exist
	secretsDir := ".secrets"
	if err := os.MkdirAll(secretsDir, 0700); err != nil {
		return fmt.Errorf("failed to create secrets directory: %w", err)
	}

	credFile := filepath.Join(secretsDir, "credentials.txt")

	// Build credentials content
	var sb strings.Builder
	sb.WriteString("# Nivo Seed Credentials\n")
	sb.WriteString(fmt.Sprintf("# Generated: %s\n", time.Now().Format(time.RFC3339)))
	sb.WriteString("# WARNING: Do not commit this file to version control\n")
	sb.WriteString("#\n")
	sb.WriteString("# Demo user credentials are public (documented in README for easy access).\n")
	sb.WriteString("# Admin credentials are generated per-instance for security.\n")
	sb.WriteString("\n")

	sb.WriteString("================================================================================\n")
	sb.WriteString("ADMIN CREDENTIALS (Generated - Not Public)\n")
	sb.WriteString("================================================================================\n")
	sb.WriteString("Email:    admin@nivo.local\n")
	sb.WriteString(fmt.Sprintf("Password: %s\n", adminPassword))
	sb.WriteString("Role:     admin\n")
	sb.WriteString("\n")

	sb.WriteString("================================================================================\n")
	sb.WriteString("DEMO USER CREDENTIALS (Public - Documented in README)\n")
	sb.WriteString("================================================================================\n")
	for _, user := range users {
		if user.Email == "admin@nivo.local" || user.Email == "admin@vnykmshr.com" {
			continue // Skip admin, already shown above
		}
		sb.WriteString(fmt.Sprintf("\n%s\n", user.FullName))
		sb.WriteString(fmt.Sprintf("  Email:    %s\n", user.Email))
		sb.WriteString(fmt.Sprintf("  Password: %s\n", user.Password))
		sb.WriteString(fmt.Sprintf("  Balance:  ₹%.2f\n", float64(user.InitialBalance)/100.0))
	}

	// Write file with restricted permissions
	if err := os.WriteFile(credFile, []byte(sb.String()), 0600); err != nil {
		return fmt.Errorf("failed to write credentials file: %w", err)
	}

	log.Printf("[%s] Credentials written to %s", serviceName, credFile)
	return nil
}

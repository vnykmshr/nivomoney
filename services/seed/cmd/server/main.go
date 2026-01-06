package main

import (
	"context"
	_ "embed"
	"encoding/json"
	"flag"
	"fmt"
	"log"
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

	log.Printf("[%s] ", serviceName)
	log.Printf("[%s] ========================================", serviceName)
	log.Printf("[%s] Seed completed successfully!", serviceName)
	log.Printf("[%s] Created/verified %d ready-to-use accounts", serviceName, len(seedData.Users))
	log.Printf("[%s] ========================================", serviceName)
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

		// Step 4: Assign 'user' role (or 'admin' for admin users)
		roleName := "user"
		if seedUser.Email == "admin@vnykmshr.com" || seedUser.Email == "admin@nivo.local" {
			roleName = "admin"
		}
		if err := assignUserRole(ctx, db, userID, roleName); err != nil {
			log.Printf("[%s]   ERROR: Failed to assign role: %v", serviceName, err)
			continue
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

		log.Printf("[%s]   ✓ Complete account ready: UserID=%s, WalletID=%s, Balance=₹%.2f",
			serviceName, userID, walletID, float64(seedUser.InitialBalance)/100.0)
	}

	log.Printf("[%s] ", serviceName)
	return nil
}

// createUser creates a user or returns existing user ID
func createUser(ctx context.Context, db *database.DB, user SeedUser) (string, error) {
	// Check if user exists
	var existingID string
	err := db.QueryRowContext(ctx, "SELECT id FROM users WHERE email = $1", user.Email).Scan(&existingID)
	if err == nil {
		log.Printf("[%s]   → User already exists: %s", serviceName, existingID)
		return existingID, nil
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}

	// Insert user
	var userID string
	query := `
		INSERT INTO users (email, phone, full_name, password_hash, status)
		VALUES ($1, $2, $3, $4, 'pending')
		RETURNING id
	`
	err = db.QueryRowContext(ctx, query, user.Email, user.Phone, user.FullName, string(hashedPassword)).Scan(&userID)
	if err != nil {
		return "", err
	}

	log.Printf("[%s]   → User created: %s", serviceName, userID)
	return userID, nil
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

	log.Printf("[%s]   → Wallet created: %s", serviceName, walletID)
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

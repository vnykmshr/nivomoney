package main

import (
	"context"
	_ "embed"
	"flag"
	"log"
	"os"
	"time"

	"gopkg.in/yaml.v3"

	identityModels "github.com/vnykmshr/nivo/services/identity/internal/models"
	identityRepo "github.com/vnykmshr/nivo/services/identity/internal/repository"
	identityService "github.com/vnykmshr/nivo/services/identity/internal/service"
	"github.com/vnykmshr/nivo/shared/config"
	"github.com/vnykmshr/nivo/shared/database"
	"github.com/vnykmshr/nivo/shared/errors"
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
	ID       string `yaml:"id"`
	FullName string `yaml:"full_name"`
	Email    string `yaml:"email"`
	Password string `yaml:"password"`
	Phone    string `yaml:"phone"`
}

func main() {
	// Parse command line flags
	cleanFlag := flag.Bool("clean", false, "Clean database before seeding")
	flag.Parse()

	log.Printf("[%s] Starting Nivo Seed Script", serviceName)
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

	// Initialize repositories
	userRepo := identityRepo.NewUserRepository(db)
	kycRepo := identityRepo.NewKYCRepository(db)
	sessionRepo := identityRepo.NewSessionRepository(db)

	// Create a minimal RBAC client (for seeding, we'll use a mock)
	rbacClient := &MockRBACClient{}

	// Create auth service (without optional dependencies for seeding)
	jwtSecret := getEnvOrDefault("JWT_SECRET", "nivo-dev-secret-change-in-production")
	authService := identityService.NewAuthService(
		userRepo,
		kycRepo,
		sessionRepo,
		rbacClient,
		nil, // wallet client - not needed for seeding
		nil, // notification client - not needed for seeding
		jwtSecret,
		24*time.Hour,
		nil, // event publisher - not needed for seeding
	)

	// Create users
	ctx := context.Background()
	userIDs := make(map[string]string)

	for i, seedUser := range seedData.Users {
		log.Printf("[%s] [%d/%d] Creating user: %s (%s)", serviceName, i+1, len(seedData.Users), seedUser.FullName, seedUser.Email)

		createReq := &identityModels.CreateUserRequest{
			Email:    seedUser.Email,
			Phone:    seedUser.Phone,
			FullName: seedUser.FullName,
			Password: seedUser.Password,
		}

		user, createErr := authService.Register(ctx, createReq)
		if createErr != nil {
			// Check if user already exists
			if createErr.Code == errors.ErrCodeConflict {
				log.Printf("[%s]   User already exists, fetching existing user", serviceName)
				existingUser, getErr := userRepo.GetByEmail(ctx, seedUser.Email)
				if getErr != nil {
					log.Printf("[%s]   ERROR: Failed to fetch existing user: %v", serviceName, getErr)
					continue
				}
				userIDs[seedUser.ID] = existingUser.ID
				log.Printf("[%s]   ✓ User exists: %s", serviceName, existingUser.ID)
			} else {
				log.Printf("[%s]   ERROR: Failed to create user: %v", serviceName, createErr)
				continue
			}
		} else {
			userIDs[seedUser.ID] = user.ID
			log.Printf("[%s]   ✓ User created: %s", serviceName, user.ID)
		}
	}

	log.Printf("[%s] ========================================", serviceName)
	log.Printf("[%s] Seed completed successfully!", serviceName)
	log.Printf("[%s] Created/verified %d users", serviceName, len(userIDs))
	log.Printf("[%s] ========================================", serviceName)
}

// cleanDatabase truncates all tables except system users
func cleanDatabase(db *database.DB) error {
	log.Printf("[%s] Cleaning database tables...", serviceName)

	// SQL to truncate all tables except system users
	cleanSQL := `
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
			risk_events,
			journal_entries,
			ledger_lines,
			accounts
		CASCADE;

		DELETE FROM users WHERE email NOT LIKE '%@vnykmshr.com';
	`

	_, err := db.Exec(cleanSQL)
	return err
}

// getEnvOrDefault returns the environment variable value or a default value
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// MockRBACClient is a mock RBAC client for seeding (doesn't make actual HTTP calls)
type MockRBACClient struct{}

func (m *MockRBACClient) AssignDefaultRole(ctx context.Context, userID string) error {
	// During seeding, we skip RBAC role assignment
	// This will be handled manually or by actual services when they're running
	log.Printf("[%s]   Skipping RBAC role assignment for user %s (seed mode)", serviceName, userID)
	return nil
}

func (m *MockRBACClient) GetUserPermissions(ctx context.Context, userID string) (*identityService.UserPermissionsResponse, error) {
	// Return empty permissions for seeding
	return &identityService.UserPermissionsResponse{
		Roles:       []identityService.RoleInfo{},
		Permissions: []identityService.Permission{},
	}, nil
}

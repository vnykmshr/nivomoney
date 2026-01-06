//nolint:gosec // G404: math/rand acceptable for simulation test data generation
package service

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"math/rand"
	"sync"
	"time"

	"github.com/vnykmshr/nivo/services/simulation/internal/behavior"
	"github.com/vnykmshr/nivo/services/simulation/internal/config"
	"github.com/vnykmshr/nivo/services/simulation/internal/metrics"
	"github.com/vnykmshr/nivo/services/simulation/internal/personas"
)

// UserWallet represents a user with their wallet
type UserWallet struct {
	UserID   string
	WalletID string
	Email    string
	Persona  personas.PersonaType
	Balance  int64 // Current balance in paise
}

// SimulationEngine generates synthetic transactions
type SimulationEngine struct {
	db               *sql.DB
	gatewayClient    *GatewayClient
	lifecycleManager *UserLifecycleManager
	users            []UserWallet     // Existing users from DB
	simulatedUsers   []*SimulatedUser // New users created by simulation

	// Thread-safe running state
	runningMu sync.RWMutex
	running   bool

	// Thread-safe random number generator
	rngMu sync.Mutex
	rng   *rand.Rand

	// Behavior injection components
	config       *config.SimulationConfig
	metrics      *metrics.SimulationMetrics
	injector     *behavior.BehaviorInjector
	autoVerifier *behavior.AutoVerifier
}

// NewSimulationEngine creates a new simulation engine
func NewSimulationEngine(
	db *sql.DB,
	gatewayClient *GatewayClient,
	cfg *config.SimulationConfig,
	met *metrics.SimulationMetrics,
) *SimulationEngine {
	injector := behavior.NewBehaviorInjector(cfg)
	autoVerifier := behavior.NewAutoVerifier(db, cfg, met)

	return &SimulationEngine{
		db:               db,
		gatewayClient:    gatewayClient,
		lifecycleManager: NewUserLifecycleManager(gatewayClient, db),
		users:            make([]UserWallet, 0),
		simulatedUsers:   make([]*SimulatedUser, 0),
		running:          false,
		rng:              rand.New(rand.NewSource(time.Now().UnixNano())),
		config:           cfg,
		metrics:          met,
		injector:         injector,
		autoVerifier:     autoVerifier,
	}
}

// randIntn returns a random int in [0, n) with thread safety.
func (s *SimulationEngine) randIntn(n int) int {
	s.rngMu.Lock()
	defer s.rngMu.Unlock()
	return s.rng.Intn(n)
}

// randFloat64 returns a random float64 in [0.0, 1.0) with thread safety.
func (s *SimulationEngine) randFloat64() float64 {
	s.rngMu.Lock()
	defer s.rngMu.Unlock()
	return s.rng.Float64()
}

// setRunning sets the running state thread-safely.
func (s *SimulationEngine) setRunning(running bool) {
	s.runningMu.Lock()
	defer s.runningMu.Unlock()
	s.running = running
}

// IsRunning returns whether the simulation is running (thread-safe).
func (s *SimulationEngine) IsRunning() bool {
	s.runningMu.RLock()
	defer s.runningMu.RUnlock()
	return s.running
}

// LoadUsers loads users and wallets from the database
func (s *SimulationEngine) LoadUsers(ctx context.Context) error {
	query := `
		SELECT
			u.id as user_id,
			w.id as wallet_id,
			u.email,
			COALESCE(w.balance, 0) as balance
		FROM users u
		INNER JOIN wallets w ON u.id = w.user_id
		LEFT JOIN user_kyc k ON u.id = k.user_id
		WHERE u.status = 'active'
			AND w.status = 'active'
			AND (k.status = 'verified' OR k.status IS NULL)
		ORDER BY u.created_at DESC
		LIMIT 50
	`

	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to query users: %w", err)
	}
	defer func() { _ = rows.Close() }()

	s.users = make([]UserWallet, 0)
	personaTypes := personas.AllPersonaTypes()

	for rows.Next() {
		var uw UserWallet
		if err := rows.Scan(&uw.UserID, &uw.WalletID, &uw.Email, &uw.Balance); err != nil {
			log.Printf("[simulation] Failed to scan user: %v", err)
			continue
		}

		// Assign random persona
		uw.Persona = personaTypes[s.randIntn(len(personaTypes))]
		s.users = append(s.users, uw)
	}

	log.Printf("[simulation] Loaded %d users for simulation", len(s.users))
	return nil
}

// Start starts the simulation engine
func (s *SimulationEngine) Start(ctx context.Context) {
	if s.IsRunning() {
		log.Printf("[simulation] Engine already running")
		return
	}

	s.setRunning(true)
	mode := "realistic"
	if s.config.IsDemo() {
		mode = "demo"
	}
	log.Printf("[simulation] Starting simulation engine (mode: %s)...", mode)

	// Load users first
	if err := s.LoadUsers(ctx); err != nil {
		log.Printf("[simulation] Failed to load users: %v", err)
		return
	}

	if len(s.users) == 0 {
		log.Printf("[simulation] No existing users found - will create simulated users")
	}

	// Create a few initial simulated users
	s.runUserCreationCycle(ctx)

	// Start auto-verification loop for simulated users
	if s.config.IsAutoVerificationEnabled() {
		go s.autoVerifier.RunAutoVerificationLoop(ctx)
	}

	// Start simulation loop
	go s.simulationLoop(ctx)
}

// Stop stops the simulation engine
func (s *SimulationEngine) Stop() {
	log.Printf("[simulation] Stopping simulation engine...")
	s.setRunning(false)
}

// simulationLoop runs the main simulation loop
func (s *SimulationEngine) simulationLoop(ctx context.Context) {
	txTicker := time.NewTicker(1 * time.Minute)        // Transaction cycle every minute
	userTicker := time.NewTicker(5 * time.Minute)      // User creation cycle every 5 minutes
	lifecycleTicker := time.NewTicker(2 * time.Minute) // Lifecycle progression every 2 minutes

	defer txTicker.Stop()
	defer userTicker.Stop()
	defer lifecycleTicker.Stop()

	log.Printf("[simulation] Simulation loop started")

	for {
		select {
		case <-ctx.Done():
			s.setRunning(false)
			return

		case <-txTicker.C:
			if !s.IsRunning() {
				return
			}
			s.runTransactionCycle(ctx)

		case <-userTicker.C:
			if !s.IsRunning() {
				return
			}
			s.runUserCreationCycle(ctx)

		case <-lifecycleTicker.C:
			if !s.IsRunning() {
				return
			}
			s.runLifecycleCycle(ctx)
		}
	}
}

// runUserCreationCycle creates new users periodically
func (s *SimulationEngine) runUserCreationCycle(ctx context.Context) {
	// Create 1-3 new users per cycle
	numUsers := s.randIntn(3) + 1
	log.Printf("[simulation] 🎭 Creating %d new users", numUsers)

	for i := 0; i < numUsers; i++ {
		user := s.lifecycleManager.GenerateNewUser()
		s.simulatedUsers = append(s.simulatedUsers, user)
		s.metrics.RecordUserCreated()
	}
}

// runLifecycleCycle progresses users through their lifecycle stages
func (s *SimulationEngine) runLifecycleCycle(ctx context.Context) {
	log.Printf("[simulation] 🔄 Running lifecycle progression")

	// Update active persona count
	activeCount := 0
	for _, user := range s.simulatedUsers {
		if user.Stage == StageActive {
			activeCount++
		}
	}
	s.metrics.SetActivePersonas(activeCount)

	for _, user := range s.simulatedUsers {
		// Apply realistic delay between operations
		if err := s.injector.ApplyDelay(ctx, "lifecycle"); err != nil {
			// Context cancelled, stop processing
			return
		}

		switch user.Stage {
		case StageNew:
			// Register the user
			if err := s.lifecycleManager.RegisterUser(ctx, user); err != nil {
				log.Printf("[simulation] Failed to register user %s: %v", user.Email, err)
				continue
			}
			// Register user for auto-verification of OTPs
			if user.UserID != "" {
				s.autoVerifier.RegisterSimulatedUser(user.UserID)
			}

		case StageRegistered:
			// Submit KYC
			if err := s.lifecycleManager.SubmitKYC(ctx, user); err != nil {
				log.Printf("[simulation] Failed to submit KYC for %s: %v", user.Email, err)
				continue
			}

		case StageKYCSubmitted:
			// Auto-verify KYC (using admin privileges)
			if err := s.lifecycleManager.VerifyKYC(ctx, user); err != nil {
				log.Printf("[simulation] Failed to verify KYC for %s: %v", user.Email, err)
				continue
			}
			s.metrics.RecordUserKYCVerified()

		case StageKYCVerified:
			// Login and mark as active
			if user.SessionToken == "" {
				if err := s.lifecycleManager.LoginUser(ctx, user); err != nil {
					log.Printf("[simulation] Failed to login user %s: %v", user.Email, err)
					continue
				}
			}
			user.Stage = StageActive
			s.metrics.RecordUserActivated()

		case StageActive:
			// Periodically re-login if session might be expired (every 12 hours)
			if time.Since(user.LastLogin) > 12*time.Hour {
				if err := s.lifecycleManager.LoginUser(ctx, user); err != nil {
					log.Printf("[simulation] Failed to re-login user %s: %v", user.Email, err)
				}
			}
		}
	}
}

// runTransactionCycle runs one cycle of transaction simulation
func (s *SimulationEngine) runTransactionCycle(ctx context.Context) {
	currentHour := time.Now().Hour()
	log.Printf("[simulation] 💸 Running transaction cycle (hour: %d)", currentHour)

	// Process transactions for existing users from DB
	for _, user := range s.users {
		persona := personas.GetPersona(user.Persona)
		if persona == nil {
			continue
		}

		// Check if user is active at this hour
		if !persona.IsActiveHour(currentHour) {
			continue
		}

		// Random chance based on frequency (simulate realistic activity)
		if !s.shouldTransact(persona.TransactionFreq) {
			continue
		}

		// Generate transaction
		if err := s.generateTransaction(ctx, user, persona); err != nil {
			log.Printf("[simulation] Failed to generate transaction for %s: %v", user.Email, err)
		}
	}

	// Process transactions for simulated users (only ACTIVE stage)
	for _, user := range s.simulatedUsers {
		if user.Stage != StageActive {
			continue
		}

		persona := personas.GetPersona(user.Persona)
		if persona == nil {
			continue
		}

		// Check if user is active at this hour
		if !persona.IsActiveHour(currentHour) {
			continue
		}

		// Random chance based on frequency
		if !s.shouldTransact(persona.TransactionFreq) {
			continue
		}

		// Generate transaction for simulated user
		if err := s.generateSimulatedUserTransaction(ctx, user, persona); err != nil {
			log.Printf("[simulation] Failed to generate transaction for simulated user %s: %v", user.Email, err)
		}
	}
}

// shouldTransact determines if a transaction should occur based on frequency
func (s *SimulationEngine) shouldTransact(freq time.Duration) bool {
	// Convert frequency to transactions per hour
	transactionsPerHour := float64(time.Hour) / float64(freq)

	// Probability that transaction occurs in this minute
	probability := transactionsPerHour / 60.0

	return s.randFloat64() < probability
}

// generateTransaction generates a single transaction based on persona
func (s *SimulationEngine) generateTransaction(ctx context.Context, user UserWallet, persona *personas.Persona) error {
	txType := persona.SelectTransactionType()
	amount := persona.RandomAmount()
	description := fmt.Sprintf("Simulated %s by %s", txType, user.Persona)

	// Check if we should inject a failure
	if s.injector.ShouldFail(txType) {
		failErr := s.injector.GetFailureError(txType)
		log.Printf("[simulation] 💥 Injected failure for %s: %s", txType, failErr.Error())
		s.metrics.RecordOperation(false, true, 0)
		return failErr
	}

	// Apply delay based on transaction type
	delayDuration := s.injector.GetDelayDuration(txType)
	if err := s.injector.ApplyDelay(ctx, txType); err != nil {
		return err
	}

	// Check balance for transfers and withdrawals
	if txType == "transfer" || txType == "withdrawal" {
		if user.Balance < amount {
			// Not enough balance - try a deposit instead
			log.Printf("[simulation] Insufficient balance for %s (balance: %d, amount: %d), creating deposit instead", txType, user.Balance, amount)
			txType = "deposit"
			// Use a smaller amount for deposit (between 1000-5000 rupees)
			amount = int64(1000+s.randIntn(4000)) * 100
		}
	}

	var err error
	switch txType {
	case "deposit":
		err = s.gatewayClient.CreateDeposit(user.WalletID, amount, description)
		if err == nil {
			// Update local balance on successful deposit
			s.updateUserBalance(user.UserID, user.Balance+amount)
		}

	case "transfer":
		// Select random recipient
		recipient := s.selectRandomUser(user.UserID)
		if recipient == nil {
			log.Printf("[simulation] No recipient available for transfer")
			return nil // Skip this transaction
		}

		err = s.gatewayClient.CreateTransfer(user.WalletID, recipient.WalletID, amount, description)
		if err == nil {
			// Update local balances on successful transfer
			s.updateUserBalance(user.UserID, user.Balance-amount)
			s.updateUserBalance(recipient.UserID, recipient.Balance+amount)
		}

	case "withdrawal":
		err = s.gatewayClient.CreateWithdrawal(user.WalletID, amount, description)
		if err == nil {
			// Update local balance on successful withdrawal
			s.updateUserBalance(user.UserID, user.Balance-amount)
		}

	default:
		return fmt.Errorf("unknown transaction type: %s", txType)
	}

	// Record metrics
	failed := err != nil
	s.metrics.RecordOperation(true, failed, delayDuration.Milliseconds())
	if !failed {
		s.metrics.RecordTransaction()
	}

	if err != nil {
		log.Printf("[simulation] Transaction failed for %s: %v", user.Email, err)
		return err
	}

	log.Printf("[simulation] ✓ %s: %s (₹%.2f) for %s", txType, description, float64(amount)/100, user.Email)
	return nil
}

// updateUserBalance updates the cached balance for a user
func (s *SimulationEngine) updateUserBalance(userID string, newBalance int64) {
	for i := range s.users {
		if s.users[i].UserID == userID {
			s.users[i].Balance = newBalance
			break
		}
	}
}

// generateSimulatedUserTransaction generates a transaction for a simulated user
func (s *SimulationEngine) generateSimulatedUserTransaction(ctx context.Context, user *SimulatedUser, persona *personas.Persona) error {
	txType := persona.SelectTransactionType()
	amount := persona.RandomAmount()
	description := fmt.Sprintf("Simulated %s by %s", txType, user.Persona)

	// Check if we should inject a failure
	if s.injector.ShouldFail(txType) {
		failErr := s.injector.GetFailureError(txType)
		log.Printf("[simulation] 💥 Injected failure for %s: %s", txType, failErr.Error())
		s.metrics.RecordOperation(false, true, 0)
		return failErr
	}

	// Apply delay based on transaction type
	delayDuration := s.injector.GetDelayDuration(txType)
	if err := s.injector.ApplyDelay(ctx, txType); err != nil {
		return err
	}

	// Check balance for transfers and withdrawals
	if txType == "transfer" || txType == "withdrawal" {
		if user.Balance < amount {
			log.Printf("[simulation] Insufficient balance for %s (balance: %d, amount: %d), creating deposit instead", txType, user.Balance, amount)
			txType = "deposit"
			amount = int64(1000+s.randIntn(4000)) * 100
		}
	}

	var err error
	switch txType {
	case "deposit":
		if user.WalletID == "" {
			log.Printf("[simulation] User %s doesn't have wallet ID yet, skipping transaction", user.Email)
			return nil
		}

		err = s.gatewayClient.CreateDeposit(user.WalletID, amount, description)
		if err == nil {
			user.Balance += amount
		}

	case "transfer":
		if user.WalletID == "" {
			log.Printf("[simulation] User %s doesn't have wallet ID yet, skipping transaction", user.Email)
			return nil
		}

		// Select random recipient (could be DB user or simulated user)
		recipient := s.selectRandomRecipient(user.UserID)
		if recipient == nil {
			log.Printf("[simulation] No recipient available for transfer")
			return nil
		}

		err = s.gatewayClient.CreateTransfer(user.WalletID, *recipient, amount, description)
		if err == nil {
			user.Balance -= amount
		}

	case "withdrawal":
		if user.WalletID == "" {
			log.Printf("[simulation] User %s doesn't have wallet ID yet, skipping transaction", user.Email)
			return nil
		}

		err = s.gatewayClient.CreateWithdrawal(user.WalletID, amount, description)
		if err == nil {
			user.Balance -= amount
		}
	}

	// Record metrics
	failed := err != nil
	s.metrics.RecordOperation(true, failed, delayDuration.Milliseconds())
	if !failed {
		s.metrics.RecordTransaction()
	}

	if err != nil {
		log.Printf("[simulation] Transaction failed for %s: %v", user.Email, err)
		return err
	}

	log.Printf("[simulation] ✓ %s: %s (₹%.2f) for %s", txType, description, float64(amount)/100, user.Email)
	return nil
}

// selectRandomUser selects a random user different from the current user
func (s *SimulationEngine) selectRandomUser(excludeUserID string) *UserWallet {
	eligible := make([]UserWallet, 0)
	for _, u := range s.users {
		if u.UserID != excludeUserID {
			eligible = append(eligible, u)
		}
	}

	if len(eligible) == 0 {
		return nil
	}

	return &eligible[s.randIntn(len(eligible))]
}

// selectRandomRecipient selects a random wallet ID for transfers (from any user pool)
func (s *SimulationEngine) selectRandomRecipient(excludeUserID string) *string {
	eligible := make([]string, 0)

	// Add DB users
	for _, u := range s.users {
		if u.UserID != excludeUserID {
			eligible = append(eligible, u.WalletID)
		}
	}

	// Add active simulated users
	for _, u := range s.simulatedUsers {
		if u.UserID != excludeUserID && u.WalletID != "" && u.Stage == StageActive {
			eligible = append(eligible, u.WalletID)
		}
	}

	if len(eligible) == 0 {
		return nil
	}

	walletID := eligible[s.randIntn(len(eligible))]
	return &walletID
}

/**
 * Shared Type Definitions
 * Common types used across user-app and admin-app
 */

// ============================================================================
// Domain Models
// ============================================================================

export interface Address {
  street: string;
  city: string;
  state: string;
  pin: string;
  country: string;
}

export interface KYCInfo {
  user_id: string;
  status: 'pending' | 'verified' | 'rejected';
  pan?: string;
  aadhaar?: string;
  date_of_birth?: string;
  address?: Address;
  verified_at?: string;
  rejected_at?: string;
  rejection_reason?: string;
  created_at: string;
  updated_at: string;
}

export interface User {
  id: string;
  email: string;
  full_name: string;
  phone: string;
  status: 'active' | 'pending' | 'suspended';
  created_at: string;
  updated_at: string;
  kyc?: KYCInfo;
}

export interface Wallet {
  id: string;
  user_id: string;
  type: 'savings' | 'current' | 'investment';
  currency: string;
  balance: number;
  available_balance: number;
  status: 'active' | 'inactive' | 'frozen' | 'closed';
  created_at: string;
  updated_at: string;
}

export interface Transaction {
  id: string;
  type: 'deposit' | 'withdrawal' | 'transfer' | 'reversal';
  status: 'pending' | 'processing' | 'completed' | 'failed' | 'reversed';
  source_wallet_id?: string;
  destination_wallet_id?: string;
  amount: number;
  currency: string;
  description: string;
  reference?: string;
  parent_transaction_id?: string;
  created_at: string;
  updated_at: string;
}

// ============================================================================
// API Request/Response Types
// ============================================================================

export interface LoginRequest {
  identifier: string; // Email or phone number
  password: string;
}

export interface RegisterRequest {
  email: string;
  password: string;
  full_name: string;
  phone: string;
}

export interface AuthResponse {
  user: User;
  token: string;
}

export interface ApiError {
  error: string;
  message: string;
  details?: Record<string, unknown>;
}

export interface ApiResponse<T> {
  success: boolean;
  data?: T;
  error?: string;
  message?: string;
}

// ============================================================================
// Transaction Request Types
// ============================================================================

export interface CreateTransferRequest {
  source_wallet_id: string;
  destination_wallet_id: string;
  amount_paise: number;
  currency: string;
  description: string;
  reference?: string;
}

export interface CreateDepositRequest {
  wallet_id: string;
  amount_paise: number;
  description: string;
  reference?: string;
}

export interface CreateWithdrawalRequest {
  wallet_id: string;
  amount_paise: number;
  description: string;
  reference?: string;
}

export interface CreateWalletRequest {
  user_id: string;
  type: 'savings' | 'current' | 'investment';
  currency: string;
}

// ============================================================================
// KYC Request Types
// ============================================================================

export interface UpdateKYCRequest {
  pan: string;
  aadhaar: string;
  date_of_birth: string;
  address: Address;
}

// ============================================================================
// Profile Request Types
// ============================================================================

export interface UpdateProfileRequest {
  full_name: string;
  email: string;
  phone: string;
}

export interface ChangePasswordRequest {
  current_password: string;
  new_password: string;
}

export interface KYCWithUser {
  kyc: KYCInfo;
  user: User;
}

// ============================================================================
// Admin Types
// ============================================================================

export interface AdminStats {
  total_users: number;
  active_users: number;
  pending_kyc: number;
  total_wallets: number;
  total_transactions: number;
}

// ============================================================================
// SSE (Server-Sent Events) Types
// ============================================================================

export interface SSEEvent {
  type: string;
  topic: string;
  data: Record<string, unknown>;
  timestamp: string;
}

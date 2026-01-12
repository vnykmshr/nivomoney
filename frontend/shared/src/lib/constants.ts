/**
 * Shared Constants
 * Common constants used across user-app and admin-app
 */

// ============================================================================
// API Configuration
// ============================================================================

export const DEFAULT_API_TIMEOUT = 15000; // 15 seconds
export const DEFAULT_API_BASE_URL = 'http://localhost:8000';

// ============================================================================
// Wallet Types
// ============================================================================

export const WALLET_TYPES = {
  SAVINGS: 'savings',
  CURRENT: 'current',
  INVESTMENT: 'investment',
} as const;

export const WALLET_TYPE_LABELS: Record<string, string> = {
  savings: 'Savings Account',
  current: 'Current Account',
  investment: 'Investment Account',
};

// ============================================================================
// Transaction Types
// ============================================================================

export const TRANSACTION_TYPES = {
  DEPOSIT: 'deposit',
  WITHDRAWAL: 'withdrawal',
  TRANSFER: 'transfer',
  REVERSAL: 'reversal',
} as const;

export const TRANSACTION_TYPE_LABELS: Record<string, string> = {
  deposit: 'Deposit',
  withdrawal: 'Withdrawal',
  transfer: 'Transfer',
  reversal: 'Reversal',
};

// ============================================================================
// Status Values
// ============================================================================

export const USER_STATUS = {
  ACTIVE: 'active',
  PENDING: 'pending',
  SUSPENDED: 'suspended',
} as const;

export const WALLET_STATUS = {
  ACTIVE: 'active',
  INACTIVE: 'inactive',
  FROZEN: 'frozen',
  CLOSED: 'closed',
} as const;

export const TRANSACTION_STATUS = {
  PENDING: 'pending',
  PROCESSING: 'processing',
  COMPLETED: 'completed',
  FAILED: 'failed',
  REVERSED: 'reversed',
  CANCELLED: 'cancelled',
} as const;

export const KYC_STATUS = {
  PENDING: 'pending',
  VERIFIED: 'verified',
  REJECTED: 'rejected',
} as const;

// ============================================================================
// Status Colors (Tailwind CSS classes)
// ============================================================================

export const STATUS_COLORS: Record<string, string> = {
  // Transaction statuses
  pending: 'bg-yellow-100 text-yellow-800',
  processing: 'bg-blue-100 text-blue-800',
  completed: 'bg-green-100 text-green-800',
  failed: 'bg-red-100 text-red-800',
  reversed: 'bg-gray-100 text-gray-800',

  // User/Wallet statuses
  active: 'bg-green-100 text-green-800',
  inactive: 'bg-gray-100 text-gray-800',
  frozen: 'bg-blue-100 text-blue-800',
  closed: 'bg-red-100 text-red-800',
  suspended: 'bg-red-100 text-red-800',

  // KYC statuses
  verified: 'bg-green-100 text-green-800',
  rejected: 'bg-red-100 text-red-800',
};

/**
 * Get status badge color classes
 */
export function getStatusColor(status: string): string {
  return STATUS_COLORS[status.toLowerCase()] || 'bg-gray-100 text-gray-800';
}

// ============================================================================
// Currency
// ============================================================================

export const DEFAULT_CURRENCY = 'INR';
export const PAISE_PER_RUPEE = 100;

// ============================================================================
// Validation Constants
// ============================================================================

export const PASSWORD_MIN_LENGTH = 8;
export const PHONE_LENGTH = 10;
export const PAN_LENGTH = 10;
export const AADHAAR_LENGTH = 12;
export const PIN_CODE_LENGTH = 6;
export const MIN_AGE = 18;

// ============================================================================
// Pagination
// ============================================================================

export const DEFAULT_PAGE_SIZE = 20;
export const MAX_PAGE_SIZE = 100;

// ============================================================================
// Session & Security
// ============================================================================

export const USER_SESSION_TIMEOUT = 24 * 60 * 60 * 1000; // 24 hours in milliseconds
export const ADMIN_SESSION_TIMEOUT = 2 * 60 * 60 * 1000; // 2 hours in milliseconds

// ============================================================================
// Storage Keys
// ============================================================================

export const STORAGE_KEYS = {
  AUTH_TOKEN: 'auth_token',
  USER_DATA: 'user_data',
  THEME: 'theme',
  LANGUAGE: 'language',
} as const;

// ============================================================================
// Error Messages
// ============================================================================

export const ERROR_MESSAGES = {
  NETWORK_ERROR: 'Network error. Please check your connection.',
  UNAUTHORIZED: 'Session expired. Please login again.',
  FORBIDDEN: 'You do not have permission to perform this action.',
  NOT_FOUND: 'The requested resource was not found.',
  SERVER_ERROR: 'Server error. Please try again later.',
  VALIDATION_ERROR: 'Please check your input and try again.',
} as const;

// ============================================================================
// Success Messages
// ============================================================================

export const SUCCESS_MESSAGES = {
  LOGIN_SUCCESS: 'Login successful!',
  REGISTER_SUCCESS: 'Registration successful!',
  KYC_SUBMITTED: 'KYC submitted successfully. We will review it shortly.',
  TRANSACTION_SUCCESS: 'Transaction completed successfully.',
  WALLET_CREATED: 'Wallet created successfully.',
} as const;

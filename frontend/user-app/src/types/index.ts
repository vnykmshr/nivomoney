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
  type: 'default';
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

export interface CreateUPIDepositRequest {
  wallet_id: string;
  amount: number; // in paise
  currency: string;
  description?: string;
}

export interface UPIDepositResponse {
  transaction: Transaction;
  virtual_upi_id: string;
  qr_code: string;
  expires_at: string;
  instructions: string[];
}

export interface CompleteUPIDepositRequest {
  transaction_id: string;
  upi_transaction_id: string;
  status: 'success' | 'failed';
}

export interface CreateWithdrawalRequest {
  wallet_id: string;
  amount_paise: number;
  description: string;
  reference?: string;
}

export interface CreateWalletRequest {
  user_id: string;
  type: 'default';
  currency: string;
}

export interface SSEEvent {
  type: string;
  topic: string;
  data: Record<string, unknown>;
  timestamp: string;
}

export interface UpdateKYCRequest {
  pan: string;
  aadhaar: string;
  date_of_birth: string;
  address: Address;
}

export interface Beneficiary {
  id: string;
  nickname: string;
  phone: string;
  wallet_id: string;
  created_at: string;
}

export interface AddBeneficiaryRequest {
  phone: string;
  nickname: string;
}

export interface UpdateBeneficiaryRequest {
  nickname: string;
}

export interface UpdateProfileRequest {
  full_name: string;
  email: string;
  phone: string;
}

export interface ChangePasswordRequest {
  current_password: string;
  new_password: string;
}

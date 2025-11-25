export interface User {
  id: string;
  email: string;
  full_name: string;
  phone: string;
  kyc_status: 'pending' | 'verified' | 'rejected';
  status: 'active' | 'inactive' | 'suspended';
  created_at: string;
  updated_at: string;
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

export interface LoginRequest {
  email: string;
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

export interface SSEEvent {
  type: string;
  topic: string;
  data: Record<string, unknown>;
  timestamp: string;
}

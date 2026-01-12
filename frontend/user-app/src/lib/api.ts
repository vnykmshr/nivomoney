import axios, { AxiosError, type AxiosInstance } from 'axios';
import {
  CSRFProtection,
  getApiBaseUrl,
  type User,
  type Wallet,
  type Transaction,
  type LoginRequest,
  type RegisterRequest,
  type AuthResponse,
  type CreateTransferRequest,
  type CreateDepositRequest,
  type CreateWithdrawalRequest,
  type CreateWalletRequest,
  type ApiError,
  type KYCInfo,
  type UpdateKYCRequest,
  type UpdateProfileRequest,
  type ChangePasswordRequest,
  type CreateUPIDepositRequest,
  type UPIDepositResponse,
  type CompleteUPIDepositRequest,
  type Beneficiary,
  type AddBeneficiaryRequest,
  type UpdateBeneficiaryRequest,
} from '@nivo/shared';

const API_BASE_URL = getApiBaseUrl();
const isProduction = import.meta.env.MODE === 'production';

class ApiClient {
  private client: AxiosInstance;
  private csrf: CSRFProtection | null = null;

  constructor() {
    this.client = axios.create({
      baseURL: API_BASE_URL,
      headers: {
        'Content-Type': 'application/json',
      },
      timeout: 15000,
    });

    // Initialize CSRF protection if in production
    if (isProduction) {
      this.csrf = new CSRFProtection();
      this.csrf.initialize();
    }

    // Add auth token and CSRF token to requests
    this.client.interceptors.request.use(config => {
      // Add auth token
      const token = localStorage.getItem('auth_token');
      if (token) {
        config.headers.Authorization = `Bearer ${token}`;
      }

      // Add CSRF token to non-GET requests
      if (this.csrf && config.method && !['get', 'head', 'options'].includes(config.method.toLowerCase())) {
        const csrfToken = this.csrf.getToken();
        if (csrfToken) {
          config.headers['X-CSRF-Token'] = csrfToken;
        }
      }

      return config;
    });

    // Handle errors globally and unwrap data
    this.client.interceptors.response.use(
      response => {
        // Unwrap the data field from API responses
        if (response.data && response.data.success && response.data.data) {
          response.data = response.data.data;
        }
        return response;
      },
      (error: AxiosError<ApiError>) => {
        if (error.response?.status === 401) {
          // Clear token and redirect to login
          localStorage.removeItem('auth_token');
          window.location.href = '/login';
        }
        return Promise.reject(error);
      }
    );
  }

  // Auth endpoints
  async login(data: LoginRequest): Promise<AuthResponse> {
    const response = await this.client.post<AuthResponse>('/api/v1/identity/auth/login', data);
    return response.data;
  }

  async register(data: RegisterRequest): Promise<AuthResponse> {
    const response = await this.client.post<AuthResponse>('/api/v1/identity/auth/register', data);
    return response.data;
  }

  async getProfile(): Promise<User> {
    const response = await this.client.get<User>('/api/v1/identity/users/me');
    return response.data;
  }

  async lookupUser(phone: string): Promise<User> {
    const response = await this.client.get<User>(`/api/v1/identity/users/lookup?phone=${encodeURIComponent(phone)}`);
    return response.data;
  }

  // Wallet endpoints
  async getWallets(): Promise<Wallet[]> {
    const response = await this.client.get<Wallet[]>('/api/v1/wallet/wallets');
    return response.data;
  }

  async getWallet(id: string): Promise<Wallet> {
    const response = await this.client.get<Wallet>(`/api/v1/wallet/wallets/${id}`);
    return response.data;
  }

  async createWallet(data: CreateWalletRequest): Promise<Wallet> {
    const response = await this.client.post<Wallet>('/api/v1/wallet/wallets', data);
    return response.data;
  }

  // Transaction endpoints
  async getTransactions(walletId: string): Promise<Transaction[]> {
    const response = await this.client.get<Transaction[]>(
      `/api/v1/transaction/wallets/${walletId}/transactions`
    );
    return response.data;
  }

  async getTransaction(id: string): Promise<Transaction> {
    const response = await this.client.get<Transaction>(`/api/v1/transaction/transactions/${id}`);
    return response.data;
  }

  async createTransfer(data: CreateTransferRequest): Promise<Transaction> {
    const response = await this.client.post<Transaction>(
      '/api/v1/transaction/transactions/transfer',
      data
    );
    return response.data;
  }

  async createDeposit(data: CreateDepositRequest): Promise<Transaction> {
    const response = await this.client.post<Transaction>(
      '/api/v1/transaction/transactions/deposit',
      data
    );
    return response.data;
  }

  async initiateUPIDeposit(data: CreateUPIDepositRequest): Promise<UPIDepositResponse> {
    const response = await this.client.post<UPIDepositResponse>(
      '/api/v1/transaction/transactions/deposit/upi',
      data
    );
    return response.data;
  }

  async completeUPIDeposit(data: CompleteUPIDepositRequest): Promise<Transaction> {
    const response = await this.client.post<Transaction>(
      '/api/v1/transaction/transactions/deposit/upi/complete',
      data
    );
    return response.data;
  }

  async createWithdrawal(data: CreateWithdrawalRequest): Promise<Transaction> {
    const response = await this.client.post<Transaction>(
      '/api/v1/transaction/transactions/withdrawal',
      data
    );
    return response.data;
  }

  // KYC endpoints
  async getKYC(): Promise<KYCInfo> {
    const response = await this.client.get<KYCInfo>('/api/v1/identity/auth/kyc');
    return response.data;
  }

  async updateKYC(data: UpdateKYCRequest): Promise<KYCInfo> {
    const response = await this.client.put<KYCInfo>('/api/v1/identity/auth/kyc', data);
    return response.data;
  }

  // Profile endpoints
  async updateProfile(data: UpdateProfileRequest): Promise<User> {
    const response = await this.client.put<User>('/api/v1/identity/users/me', data);
    return response.data;
  }

  async changePassword(data: ChangePasswordRequest): Promise<void> {
    await this.client.put('/api/v1/identity/users/me/password', data);
  }

  // Beneficiary endpoints
  async getBeneficiaries(): Promise<Beneficiary[]> {
    const response = await this.client.get<Beneficiary[]>('/api/v1/wallet/beneficiaries');
    return response.data;
  }

  async addBeneficiary(data: AddBeneficiaryRequest): Promise<Beneficiary> {
    const response = await this.client.post<Beneficiary>('/api/v1/wallet/beneficiaries', data);
    return response.data;
  }

  async updateBeneficiary(id: string, data: UpdateBeneficiaryRequest): Promise<Beneficiary> {
    const response = await this.client.put<Beneficiary>(`/api/v1/wallet/beneficiaries/${id}`, data);
    return response.data;
  }

  async deleteBeneficiary(id: string): Promise<void> {
    await this.client.delete(`/api/v1/wallet/beneficiaries/${id}`);
  }

  // Note: Admin endpoints have been moved to admin-app
  // See: /frontend/admin-app/src/lib/adminApi.ts

  /**
   * Clear CSRF token (call on logout)
   */
  clearCSRF(): void {
    if (this.csrf) {
      this.csrf.clear();
    }
  }

  /**
   * Regenerate CSRF token (call on login)
   */
  regenerateCSRF(): void {
    if (this.csrf) {
      this.csrf.setToken(this.csrf.generateToken());
    }
  }
}

export const api = new ApiClient();

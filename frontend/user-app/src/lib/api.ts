import axios, { AxiosError, type AxiosInstance } from 'axios';
import type {
  User,
  Wallet,
  Transaction,
  LoginRequest,
  RegisterRequest,
  AuthResponse,
  CreateTransferRequest,
  CreateDepositRequest,
  CreateWithdrawalRequest,
  CreateWalletRequest,
  ApiError,
} from '../types';

const API_BASE_URL = import.meta.env.VITE_API_URL || 'http://localhost:8000';

class ApiClient {
  private client: AxiosInstance;

  constructor() {
    this.client = axios.create({
      baseURL: API_BASE_URL,
      headers: {
        'Content-Type': 'application/json',
      },
      timeout: 15000,
    });

    // Add auth token to requests
    this.client.interceptors.request.use(config => {
      const token = localStorage.getItem('auth_token');
      if (token) {
        config.headers.Authorization = `Bearer ${token}`;
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
      `/api/v1/transaction/transactions/wallet/${walletId}`
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

  async createWithdrawal(data: CreateWithdrawalRequest): Promise<Transaction> {
    const response = await this.client.post<Transaction>(
      '/api/v1/transaction/transactions/withdrawal',
      data
    );
    return response.data;
  }
}

export const api = new ApiClient();

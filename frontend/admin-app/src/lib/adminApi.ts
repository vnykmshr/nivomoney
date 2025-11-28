/**
 * Admin API Client
 * Extends BaseApiClient from @nivo/shared with admin-specific endpoints
 */

import {
  BaseApiClient,
  type User,
  type KYCWithUser,
  type AdminStats,
  type LoginRequest,
  type AuthResponse,
  type Wallet,
  type FreezeWalletRequest,
  type CloseWalletRequest,
  type Transaction,
} from '@nivo/shared';

const API_BASE_URL = import.meta.env.VITE_API_URL || 'http://localhost:8000';

// Determine environment for CSRF protection
const isProduction = import.meta.env.MODE === 'production';

class AdminApiClient extends BaseApiClient {
  constructor() {
    super({
      baseURL: API_BASE_URL,
      timeout: 15000,
      getToken: () => localStorage.getItem('admin_token'),
      onUnauthorized: () => {
        localStorage.removeItem('admin_token');
        window.location.href = '/login';
      },
      csrfEnabled: isProduction, // Enable CSRF in production only
      environment: isProduction ? 'production' : 'development',
    });
  }

  // ============================================================================
  // Authentication Endpoints
  // ============================================================================

  async login(data: LoginRequest): Promise<AuthResponse> {
    const response = await this.post<AuthResponse>('/api/v1/identity/auth/login', data);
    return response;
  }

  async logout(): Promise<void> {
    await this.post('/api/v1/identity/auth/logout');
  }

  async getProfile(): Promise<User> {
    const response = await this.get<User>('/api/v1/identity/users/me');
    return response;
  }

  // ============================================================================
  // Admin Statistics Endpoints
  // ============================================================================

  async getAdminStats(): Promise<AdminStats> {
    const response = await this.get<AdminStats>('/api/v1/identity/admin/stats');
    return response;
  }

  // ============================================================================
  // KYC Management Endpoints
  // ============================================================================

  async listPendingKYCs(limit = 50, offset = 0): Promise<KYCWithUser[]> {
    const response = await this.get<KYCWithUser[]>(
      `/api/v1/identity/admin/kyc/pending?limit=${limit}&offset=${offset}`
    );
    return response;
  }

  async verifyKYC(userId: string): Promise<void> {
    await this.post('/api/v1/identity/admin/kyc/verify', { user_id: userId });
  }

  async rejectKYC(userId: string, reason: string): Promise<void> {
    await this.post('/api/v1/identity/admin/kyc/reject', { user_id: userId, reason });
  }

  // ============================================================================
  // User Management Endpoints (Future)
  // ============================================================================

  async searchUsers(query: string, limit = 50, offset = 0): Promise<User[]> {
    const response = await this.get<User[]>(
      `/api/v1/identity/admin/users/search?q=${encodeURIComponent(query)}&limit=${limit}&offset=${offset}`
    );
    return response;
  }

  async getUserDetails(userId: string): Promise<User> {
    const response = await this.get<User>(`/api/v1/identity/admin/users/${userId}`);
    return response;
  }

  async updateUser(_userId: string, _data: Partial<User>): Promise<User> {
    // TODO: Implement when backend endpoint is ready
    throw new Error('User update endpoint not yet implemented');
  }

  async suspendUser(_userId: string, _reason: string): Promise<void> {
    // TODO: Implement when backend endpoint is ready
    throw new Error('User suspension endpoint not yet implemented');
  }

  // ============================================================================
  // Transaction Details Endpoints (Future)
  // ============================================================================

  async getTransactionDetails(_txId: string): Promise<any> {
    // TODO: Implement when backend endpoint is ready
    throw new Error('Transaction details endpoint not yet implemented');
  }

  // ============================================================================
  // Wallet Management Endpoints
  // ============================================================================

  async getUserWallets(userId: string, status?: string): Promise<Wallet[]> {
    let url = `/api/v1/wallet/users/${userId}/wallets`;
    if (status) {
      url += `?status=${encodeURIComponent(status)}`;
    }
    const response = await this.get<Wallet[]>(url);
    return response;
  }

  async freezeWallet(walletId: string, data: FreezeWalletRequest): Promise<Wallet> {
    const response = await this.post<Wallet>(`/api/v1/wallet/wallets/${walletId}/freeze`, data);
    return response;
  }

  async unfreezeWallet(walletId: string): Promise<Wallet> {
    const response = await this.post<Wallet>(`/api/v1/wallet/wallets/${walletId}/unfreeze`, {});
    return response;
  }

  async closeWallet(walletId: string, data: CloseWalletRequest): Promise<Wallet> {
    const response = await this.post<Wallet>(`/api/v1/wallet/wallets/${walletId}/close`, data);
    return response;
  }

  // ============================================================================
  // Transaction Monitoring Endpoints
  // ============================================================================

  async searchTransactions(filters: {
    transaction_id?: string;
    user_id?: string;
    status?: string;
    type?: string;
    search?: string;
    min_amount?: number;
    max_amount?: number;
    limit?: number;
    offset?: number;
  }): Promise<Transaction[]> {
    const params = new URLSearchParams();
    Object.entries(filters).forEach(([key, value]) => {
      if (value !== undefined && value !== '') {
        params.append(key, String(value));
      }
    });

    const response = await this.get<Transaction[]>(
      `/api/v1/admin/transactions/search?${params.toString()}`
    );
    return response;
  }
}

export const adminApi = new AdminApiClient();

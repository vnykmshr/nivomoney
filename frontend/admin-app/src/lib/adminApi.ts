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

  async searchUsers(_query: string): Promise<User[]> {
    // TODO: Implement when backend endpoint is ready
    throw new Error('User search endpoint not yet implemented');
  }

  async getUserDetails(_userId: string): Promise<User> {
    // TODO: Implement when backend endpoint is ready
    throw new Error('User details endpoint not yet implemented');
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
  // Transaction Management Endpoints (Future)
  // ============================================================================

  async searchTransactions(_query: string): Promise<any[]> {
    // TODO: Implement when backend endpoint is ready
    throw new Error('Transaction search endpoint not yet implemented');
  }

  async getTransactionDetails(_txId: string): Promise<any> {
    // TODO: Implement when backend endpoint is ready
    throw new Error('Transaction details endpoint not yet implemented');
  }
}

export const adminApi = new AdminApiClient();

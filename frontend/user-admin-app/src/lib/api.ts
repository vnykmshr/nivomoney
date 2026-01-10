/**
 * User-Admin API Client
 * API client for User-Admin verification portal
 */

import { BaseApiClient, type User, type LoginRequest, type AuthResponse } from '@nivo/shared';

const API_BASE_URL = import.meta.env.VITE_API_URL || 'http://localhost:8000';

// Verification request type
export interface Verification {
  id: string;
  user_id: string;
  operation_type: string;
  status: 'pending' | 'verified' | 'expired' | 'cancelled';
  otp_code?: string; // Only visible to User-Admin
  expires_at: string;
  created_at: string;
  verified_at?: string;
  metadata?: Record<string, unknown>;
}

// Paired user profile
export interface PairedUser {
  id: string;
  full_name: string;
  email: string;
  phone: string;
  kyc_status: string;
}

class UserAdminApiClient extends BaseApiClient {
  constructor() {
    super({
      baseURL: API_BASE_URL,
      timeout: 15000,
      getToken: () => localStorage.getItem('user_admin_token'),
      onUnauthorized: () => {
        localStorage.removeItem('user_admin_token');
        window.location.href = '/login';
      },
      csrfEnabled: false, // Not needed for verification portal
      environment: import.meta.env.MODE === 'production' ? 'production' : 'development',
    });
  }

  // ============================================================================
  // Authentication Endpoints
  // ============================================================================

  async login(data: LoginRequest): Promise<AuthResponse> {
    const response = await this.post<AuthResponse>('/api/v1/identity/auth/login', {
      ...data,
      account_type: 'user_admin', // Specify user-admin account type
    });
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
  // Verification Endpoints
  // ============================================================================

  async getPendingVerifications(): Promise<Verification[]> {
    const response = await this.get<Verification[]>('/api/v1/identity/verifications/pending');
    return response;
  }

  async getVerification(id: string): Promise<Verification> {
    const response = await this.get<Verification>(`/api/v1/identity/verifications/${id}`);
    return response;
  }

  // ============================================================================
  // Paired User Endpoints
  // ============================================================================

  async getPairedUser(): Promise<PairedUser> {
    const response = await this.get<PairedUser>('/api/v1/identity/user-admin/paired-user');
    return response;
  }
}

export const api = new UserAdminApiClient();

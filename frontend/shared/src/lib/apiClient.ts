/**
 * Base API Client
 * Shared HTTP client configuration for both user-app and admin-app
 */

import axios, { AxiosError, AxiosInstance, AxiosRequestConfig, AxiosResponse } from 'axios';
import type { ApiError } from '../types';
import { CSRFProtection, type Environment } from './security';

export interface ApiClientConfig {
  baseURL: string;
  timeout?: number;
  getToken?: () => string | null;
  onUnauthorized?: () => void;
  csrfEnabled?: boolean;
  environment?: Environment;
}

/**
 * Base API Client class
 * Provides common functionality for HTTP requests with authentication
 */
export class BaseApiClient {
  protected client: AxiosInstance;
  protected config: ApiClientConfig;
  protected csrf: CSRFProtection | null = null;

  constructor(config: ApiClientConfig) {
    this.config = config;
    this.client = axios.create({
      baseURL: config.baseURL,
      headers: {
        'Content-Type': 'application/json',
      },
      timeout: config.timeout || 15000,
    });

    // Initialize CSRF protection if enabled
    if (config.csrfEnabled) {
      this.csrf = new CSRFProtection();
      this.csrf.initialize();
    }

    this.setupInterceptors();
  }

  /**
   * Set up request and response interceptors
   */
  private setupInterceptors(): void {
    // Request interceptor: Add auth token and CSRF token
    this.client.interceptors.request.use(
      (config) => {
        // Add auth token
        if (this.config.getToken) {
          const token = this.config.getToken();
          if (token) {
            config.headers.Authorization = `Bearer ${token}`;
          }
        }

        // Add CSRF token to non-GET requests
        if (this.csrf && config.method && !['get', 'head', 'options'].includes(config.method.toLowerCase())) {
          const csrfToken = this.csrf.getToken();
          if (csrfToken) {
            config.headers['X-CSRF-Token'] = csrfToken;
          }
        }

        return config;
      },
      (error) => Promise.reject(error)
    );

    // Response interceptor: Unwrap data and handle errors
    this.client.interceptors.response.use(
      (response: AxiosResponse) => {
        // Unwrap the data field from API responses
        if (response.data && response.data.success && response.data.data !== undefined) {
          response.data = response.data.data;
        }
        return response;
      },
      (error: AxiosError<ApiError>) => {
        // Handle 401 Unauthorized
        if (error.response?.status === 401 && this.config.onUnauthorized) {
          this.config.onUnauthorized();
        }
        return Promise.reject(error);
      }
    );
  }

  /**
   * GET request
   */
  protected async get<T>(url: string, config?: AxiosRequestConfig): Promise<T> {
    const response = await this.client.get<T>(url, config);
    return response.data;
  }

  /**
   * POST request
   */
  protected async post<T>(url: string, data?: unknown, config?: AxiosRequestConfig): Promise<T> {
    const response = await this.client.post<T>(url, data, config);
    return response.data;
  }

  /**
   * PUT request
   */
  protected async put<T>(url: string, data?: unknown, config?: AxiosRequestConfig): Promise<T> {
    const response = await this.client.put<T>(url, data, config);
    return response.data;
  }

  /**
   * DELETE request
   */
  protected async delete<T>(url: string, config?: AxiosRequestConfig): Promise<T> {
    const response = await this.client.delete<T>(url, config);
    return response.data;
  }

  /**
   * PATCH request
   */
  protected async patch<T>(url: string, data?: unknown, config?: AxiosRequestConfig): Promise<T> {
    const response = await this.client.patch<T>(url, data, config);
    return response.data;
  }

  /**
   * Get the underlying Axios instance (for advanced usage)
   */
  public getAxiosInstance(): AxiosInstance {
    return this.client;
  }

  /**
   * Clear CSRF token (call on logout)
   */
  public clearCSRF(): void {
    if (this.csrf) {
      this.csrf.clear();
    }
  }

  /**
   * Regenerate CSRF token (call on login)
   */
  public regenerateCSRF(): void {
    if (this.csrf) {
      this.csrf.setToken(this.csrf.generateToken());
    }
  }
}

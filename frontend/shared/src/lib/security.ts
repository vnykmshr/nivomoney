/**
 * Security Configuration and Utilities
 * Shared security settings and functions for both user-app and admin-app
 */

// Environment type
export type Environment = 'development' | 'production';

// Environment detection helpers (to be used in apps, not in shared package)
export const isProduction = (env: Environment): boolean => {
  return env === 'production';
};

export const isDevelopment = (env: Environment): boolean => {
  return env === 'development';
};

// Security Configuration
export interface SecurityConfig {
  // Session settings
  sessionTimeout: number; // milliseconds
  sessionCheckInterval: number; // milliseconds

  // CSRF settings
  csrfEnabled: boolean;
  csrfTokenKey: string;

  // Content Security Policy
  cspEnabled: boolean;

  // HTTP Security Headers
  securityHeaders: {
    strictTransportSecurity: boolean;
    xFrameOptions: boolean;
    xContentTypeOptions: boolean;
    referrerPolicy: boolean;
  };

  // IP Whitelisting (production only)
  ipWhitelistEnabled: boolean;
  allowedIPs: string[];

  // Audit logging
  auditLoggingEnabled: boolean;
}

// User App Security Config
export const userAppSecurityConfig: Record<string, SecurityConfig> = {
  development: {
    sessionTimeout: 24 * 60 * 60 * 1000, // 24 hours
    sessionCheckInterval: 5 * 60 * 1000, // 5 minutes
    csrfEnabled: false, // Disabled in dev for easier testing
    csrfTokenKey: 'X-CSRF-Token',
    cspEnabled: false,
    securityHeaders: {
      strictTransportSecurity: false,
      xFrameOptions: true,
      xContentTypeOptions: true,
      referrerPolicy: true,
    },
    ipWhitelistEnabled: false,
    allowedIPs: [],
    auditLoggingEnabled: false,
  },
  production: {
    sessionTimeout: 24 * 60 * 60 * 1000, // 24 hours
    sessionCheckInterval: 5 * 60 * 1000, // 5 minutes
    csrfEnabled: true,
    csrfTokenKey: 'X-CSRF-Token',
    cspEnabled: true,
    securityHeaders: {
      strictTransportSecurity: true,
      xFrameOptions: true,
      xContentTypeOptions: true,
      referrerPolicy: true,
    },
    ipWhitelistEnabled: false, // Public app
    allowedIPs: [],
    auditLoggingEnabled: true,
  },
};

// Admin App Security Config (Stricter)
export const adminAppSecurityConfig: Record<string, SecurityConfig> = {
  development: {
    sessionTimeout: 2 * 60 * 60 * 1000, // 2 hours
    sessionCheckInterval: 1 * 60 * 1000, // 1 minute
    csrfEnabled: false, // Disabled in dev
    csrfTokenKey: 'X-CSRF-Token',
    cspEnabled: false,
    securityHeaders: {
      strictTransportSecurity: false,
      xFrameOptions: true,
      xContentTypeOptions: true,
      referrerPolicy: true,
    },
    ipWhitelistEnabled: false,
    allowedIPs: [],
    auditLoggingEnabled: true, // Always enabled for admin
  },
  production: {
    sessionTimeout: 2 * 60 * 60 * 1000, // 2 hours
    sessionCheckInterval: 1 * 60 * 1000, // 1 minute
    csrfEnabled: true,
    csrfTokenKey: 'X-CSRF-Token',
    cspEnabled: true,
    securityHeaders: {
      strictTransportSecurity: true,
      xFrameOptions: true,
      xContentTypeOptions: true,
      referrerPolicy: true,
    },
    ipWhitelistEnabled: true, // VPN/IP restricted in production
    allowedIPs: [
      // Add your office/VPN IPs here
      // '203.0.113.0/24',  // Example: Office network
      // '198.51.100.50',   // Example: VPN gateway
    ],
    auditLoggingEnabled: true,
  },
};

// Get current security config
export const getSecurityConfig = (isAdminApp: boolean, env: Environment = 'development'): SecurityConfig => {
  return isAdminApp
    ? adminAppSecurityConfig[env]
    : userAppSecurityConfig[env];
};

// CSRF Token Management
export class CSRFProtection {
  private tokenKey: string;

  constructor(tokenKey: string = 'X-CSRF-Token') {
    this.tokenKey = tokenKey;
  }

  /**
   * Generate a CSRF token (to be stored in sessionStorage)
   */
  generateToken(): string {
    const array = new Uint8Array(32);
    crypto.getRandomValues(array);
    return Array.from(array, byte => byte.toString(16).padStart(2, '0')).join('');
  }

  /**
   * Get CSRF token from storage
   */
  getToken(): string | null {
    return sessionStorage.getItem(this.tokenKey);
  }

  /**
   * Set CSRF token in storage
   */
  setToken(token: string): void {
    sessionStorage.setItem(this.tokenKey, token);
  }

  /**
   * Initialize CSRF token (call on app start)
   */
  initialize(): void {
    if (!this.getToken()) {
      this.setToken(this.generateToken());
    }
  }

  /**
   * Clear CSRF token (call on logout)
   */
  clear(): void {
    sessionStorage.removeItem(this.tokenKey);
  }
}

// Input Sanitization (XSS Prevention)
export const sanitizeHTML = (input: string): string => {
  const div = document.createElement('div');
  div.textContent = input;
  return div.innerHTML;
};

export const sanitizeURL = (url: string): string => {
  try {
    const parsed = new URL(url);
    // Only allow http and https protocols
    if (!['http:', 'https:'].includes(parsed.protocol)) {
      return '';
    }
    return parsed.toString();
  } catch {
    return '';
  }
};

// Content Security Policy
export const getCSPDirectives = (isAdminApp: boolean, env: Environment = 'development'): string => {
  const baseDirectives = [
    "default-src 'self'",
    "script-src 'self' 'unsafe-inline'", // Note: unsafe-inline needed for Vite
    "style-src 'self' 'unsafe-inline'",
    "img-src 'self' data: https:",
    "font-src 'self' data:",
    "connect-src 'self' http://localhost:* ws://localhost:*", // Allow local API
  ];

  if (env === 'production') {
    // In production, tighten up
    return baseDirectives
      .map(d => d.replace(/localhost:\*/g, isAdminApp ? 'admin.nivomoney.com' : 'app.nivomoney.com'))
      .join('; ');
  }

  return baseDirectives.join('; ');
};

// Security Headers (for Vite config)
export const getSecurityHeaders = (config: SecurityConfig): Record<string, string> => {
  const headers: Record<string, string> = {};

  if (config.securityHeaders.strictTransportSecurity) {
    headers['Strict-Transport-Security'] = 'max-age=31536000; includeSubDomains';
  }

  if (config.securityHeaders.xFrameOptions) {
    headers['X-Frame-Options'] = 'DENY';
  }

  if (config.securityHeaders.xContentTypeOptions) {
    headers['X-Content-Type-Options'] = 'nosniff';
  }

  if (config.securityHeaders.referrerPolicy) {
    headers['Referrer-Policy'] = 'strict-origin-when-cross-origin';
  }

  if (config.cspEnabled) {
    headers['Content-Security-Policy'] = getCSPDirectives(false); // Set per app
  }

  return headers;
};

// Rate Limiting (client-side tracking)
export class ClientRateLimiter {
  private requests: Map<string, number[]> = new Map();

  /**
   * Check if request is allowed
   * @param key - Unique key for the request (e.g., endpoint name)
   * @param maxRequests - Maximum requests allowed
   * @param windowMs - Time window in milliseconds
   */
  isAllowed(key: string, maxRequests: number, windowMs: number): boolean {
    const now = Date.now();
    const timestamps = this.requests.get(key) || [];

    // Remove old timestamps outside the window
    const validTimestamps = timestamps.filter(ts => now - ts < windowMs);

    // Check if limit exceeded
    if (validTimestamps.length >= maxRequests) {
      return false;
    }

    // Add current timestamp
    validTimestamps.push(now);
    this.requests.set(key, validTimestamps);

    return true;
  }

  /**
   * Reset rate limit for a key
   */
  reset(key: string): void {
    this.requests.delete(key);
  }

  /**
   * Clear all rate limits
   */
  clearAll(): void {
    this.requests.clear();
  }
}

// Audit Logging
export interface AuditLog {
  timestamp: string;
  userId: string;
  userName: string;
  action: string;
  resource: string;
  details?: Record<string, unknown>;
  ipAddress?: string;
  userAgent?: string;
}

export class AuditLogger {
  private logs: AuditLog[] = [];
  private maxLogs: number = 1000;
  private enabled: boolean;
  private env: Environment;

  constructor(enabled: boolean = true, env: Environment = 'development') {
    this.enabled = enabled;
    this.env = env;
  }

  /**
   * Log an admin action
   */
  log(log: Omit<AuditLog, 'timestamp'>): void {
    if (!this.enabled) return;

    const entry: AuditLog = {
      ...log,
      timestamp: new Date().toISOString(),
    };

    // Add to in-memory logs
    this.logs.push(entry);

    // Keep only last N logs
    if (this.logs.length > this.maxLogs) {
      this.logs.shift();
    }

    // Log to console in development
    if (this.env === 'development') {
      console.log('[AUDIT]', entry);
    }

    // In production, send to backend
    if (this.env === 'production') {
      this.sendToBackend(entry);
    }
  }

  /**
   * Send audit log to backend
   */
  private async sendToBackend(_log: AuditLog): Promise<void> {
    try {
      // TODO: Implement backend audit log endpoint
      // await fetch('/api/v1/admin/audit-log', {
      //   method: 'POST',
      //   headers: { 'Content-Type': 'application/json' },
      //   body: JSON.stringify(log),
      // });
    } catch (error) {
      console.error('Failed to send audit log:', error);
    }
  }

  /**
   * Get all logs
   */
  getLogs(): AuditLog[] {
    return [...this.logs];
  }

  /**
   * Clear all logs
   */
  clearLogs(): void {
    this.logs = [];
  }
}

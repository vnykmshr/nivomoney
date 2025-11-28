/**
 * Admin Login Page
 * Secure login page for admin access
 * Features: Stricter validation, no "remember me", session timeout warnings
 */

import { useState, type FormEvent } from 'react';
import { useNavigate, useLocation } from 'react-router-dom';
import { useAdminAuthStore } from '../stores/adminAuthStore';

export function AdminLogin() {
  const navigate = useNavigate();
  const location = useLocation();
  const { login, error, clearError } = useAdminAuthStore();

  const [identifier, setIdentifier] = useState('');
  const [password, setPassword] = useState('');
  const [isLoading, setIsLoading] = useState(false);

  const from = (location.state as any)?.from?.pathname || '/';

  const handleSubmit = async (e: FormEvent) => {
    e.preventDefault();
    clearError();

    if (!identifier || !password) {
      return;
    }

    try {
      setIsLoading(true);
      await login(identifier, password);
      navigate(from, { replace: true });
    } catch (error) {
      // Error is handled by the store
      console.error('Login failed:', error);
    } finally {
      setIsLoading(false);
    }
  };

  return (
    <div className="min-h-screen bg-gradient-to-br from-primary-50 to-secondary-100 flex items-center justify-center p-4">
      <div className="max-w-md w-full">
        {/* Admin Badge */}
        <div className="text-center mb-6">
          <div className="inline-block px-4 py-2 bg-accent-100 text-accent-800 rounded-full font-medium mb-4">
            🔐 Admin Access
          </div>
          <h1 className="text-3xl font-bold text-gray-900 mb-2">Nivo Money Admin</h1>
          <p className="text-gray-600">Secure administrative access</p>
        </div>

        {/* Login Card */}
        <div className="card">
          {/* Security Notice */}
          <div className="mb-6 p-3 bg-amber-50 border-l-4 border-amber-400 text-sm">
            <p className="text-amber-800">
              <strong>Security Notice:</strong> Admin sessions expire after 2 hours of inactivity.
              All admin actions are logged and monitored.
            </p>
          </div>

          {/* Error Message */}
          {error && (
            <div className="mb-4 p-3 bg-red-100 text-red-800 rounded-lg flex justify-between items-center">
              <span>{error}</span>
              <button onClick={clearError} className="text-red-600 hover:text-red-800">
                ✕
              </button>
            </div>
          )}

          {/* Login Form */}
          <form onSubmit={handleSubmit} className="space-y-4">
            <div>
              <label htmlFor="identifier" className="block text-sm font-medium text-gray-700 mb-1">
                Email or Phone
              </label>
              <input
                id="identifier"
                type="text"
                value={identifier}
                onChange={(e) => setIdentifier(e.target.value)}
                placeholder="admin@nivomoney.com"
                className="input-field"
                required
                autoComplete="username"
              />
            </div>

            <div>
              <label htmlFor="password" className="block text-sm font-medium text-gray-700 mb-1">
                Password
              </label>
              <input
                id="password"
                type="password"
                value={password}
                onChange={(e) => setPassword(e.target.value)}
                placeholder="••••••••"
                className="input-field"
                required
                autoComplete="current-password"
              />
            </div>

            <button
              type="submit"
              disabled={isLoading || !identifier || !password}
              className="w-full btn-primary py-3"
            >
              {isLoading ? 'Authenticating...' : 'Admin Login'}
            </button>
          </form>

          {/* Security Footer */}
          <div className="mt-6 pt-6 border-t border-gray-200">
            <div className="flex items-start space-x-2 text-xs text-gray-600">
              <span>🔒</span>
              <div>
                <p className="font-medium mb-1">Security Features:</p>
                <ul className="list-disc list-inside space-y-1">
                  <li>2-hour session timeout</li>
                  <li>Activity-based auto-logout</li>
                  <li>Audit logging enabled</li>
                  <li>VPN required in production</li>
                </ul>
              </div>
            </div>
          </div>
        </div>

        {/* Help Text */}
        <div className="mt-6 text-center text-sm text-gray-600">
          <p>Admin credentials required. Contact IT support for access.</p>
        </div>
      </div>
    </div>
  );
}

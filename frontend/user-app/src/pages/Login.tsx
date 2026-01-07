import { useState } from 'react';
import { useNavigate, Link } from 'react-router-dom';
import { useAuthStore } from '../stores/authStore';
import { normalizeIndianPhone } from '../lib/utils';
import {
  LogoWithText,
  Card,
  Button,
  Input,
  FormField,
  Alert,
} from '../../../shared/components';

export function Login() {
  const [identifier, setIdentifier] = useState('');
  const [password, setPassword] = useState('');
  const [errors, setErrors] = useState<Record<string, string>>({});
  const navigate = useNavigate();

  const { login, isLoading, error: authError } = useAuthStore();

  const validate = (): boolean => {
    const newErrors: Record<string, string> = {};

    if (!identifier) {
      newErrors.identifier = 'Email or phone is required';
    }

    if (!password) {
      newErrors.password = 'Password is required';
    } else if (password.length < 6) {
      newErrors.password = 'Password must be at least 6 characters';
    }

    setErrors(newErrors);
    return Object.keys(newErrors).length === 0;
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();

    if (!validate()) return;

    try {
      const normalizedIdentifier = normalizeIndianPhone(identifier);
      await login(normalizedIdentifier, password);
      navigate('/dashboard');
    } catch (err) {
      console.error('Login failed:', err);
    }
  };

  return (
    <div className="min-h-screen flex items-center justify-center bg-[var(--surface-page)] px-4">
      <div className="w-full max-w-md">
        {/* Logo */}
        <div className="text-center mb-8">
          <LogoWithText className="justify-center" />
          <p className="mt-2 text-[var(--text-secondary)]">
            Sign in to your account
          </p>
        </div>

        <Card padding="lg">
          <form onSubmit={handleSubmit} className="space-y-5">
            {authError && (
              <Alert variant="error">
                {authError}
              </Alert>
            )}

            <FormField
              label="Email or Phone"
              htmlFor="identifier"
              error={errors.identifier}
              required
            >
              <Input
                id="identifier"
                type="text"
                value={identifier}
                onChange={e => setIdentifier(e.target.value)}
                placeholder="you@example.com or 9876543210"
                disabled={isLoading}
                error={!!errors.identifier}
                autoComplete="username"
              />
            </FormField>

            <FormField
              label="Password"
              htmlFor="password"
              error={errors.password}
              required
            >
              <Input
                id="password"
                type="password"
                value={password}
                onChange={e => setPassword(e.target.value)}
                placeholder="Enter your password"
                disabled={isLoading}
                error={!!errors.password}
                autoComplete="current-password"
              />
            </FormField>

            <Button
              type="submit"
              className="w-full"
              loading={isLoading}
              size="lg"
            >
              Sign In
            </Button>
          </form>

          <div className="mt-6 text-center">
            <p className="text-sm text-[var(--text-secondary)]">
              Don't have an account?{' '}
              <Link
                to="/register"
                className="text-[var(--text-link)] hover:text-[var(--text-link-hover)] font-medium"
              >
                Sign up
              </Link>
            </p>
          </div>
        </Card>

        {/* Footer */}
        <p className="mt-8 text-center text-xs text-[var(--text-muted)]">
          By signing in, you agree to our Terms of Service and Privacy Policy.
        </p>
      </div>
    </div>
  );
}

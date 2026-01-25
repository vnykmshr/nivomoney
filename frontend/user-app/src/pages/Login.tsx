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
  PageHero,
  TrustBadge,
} from '@nivo/shared';

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
    } catch {
      // Error is already handled by authStore
    }
  };

  return (
    <main id="main-content" tabIndex={-1} className="min-h-screen flex flex-col bg-[var(--surface-page)] outline-none">
      {/* Dark Hero Section with integrated wave */}
      <PageHero variant="dark" size="md" showGlow showGrid showWave>
        <div className="text-center">
          <LogoWithText className="justify-center" size="lg" variant="light" />
          <p className="mt-4 text-lg text-neutral-300 max-w-md mx-auto">
            Your trusted partner for seamless money transfers
          </p>

          {/* Trust Badges */}
          <div className="flex flex-wrap justify-center gap-3 mt-6">
            <TrustBadge variant="security" size="sm" theme="dark" />
            <TrustBadge variant="encrypted" size="sm" theme="dark" />
          </div>
        </div>
      </PageHero>

      {/* Form Section - overlaps the wave */}
      <div className="flex-1 px-4 -mt-8 relative z-10 pb-8">
        <div className="w-full max-w-md mx-auto">
          <Card padding="lg" variant="elevated" className="shadow-xl">
            <div className="text-center mb-6">
              <h1 className="text-2xl font-bold text-[var(--text-primary)]">
                Welcome Back
              </h1>
              <p className="mt-1 text-[var(--text-secondary)]">
                Sign in to your account
              </p>
            </div>

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
          <p className="mt-6 text-center text-xs text-[var(--text-muted)]">
            By signing in, you agree to our{' '}
            <Link to="/terms" className="underline hover:text-[var(--text-secondary)]">
              Terms of Service
            </Link>{' '}
            and{' '}
            <Link to="/privacy" className="underline hover:text-[var(--text-secondary)]">
              Privacy Policy
            </Link>
            .
          </p>
        </div>
      </div>
    </main>
  );
}

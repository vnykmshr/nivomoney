import { Link } from 'react-router-dom';
import { LogoWithText, Card, Button, PageHero } from '@nivo/shared';

export function Terms() {
  return (
    <main id="main-content" tabIndex={-1} className="min-h-screen flex flex-col bg-[var(--surface-page)] outline-none">
      <PageHero variant="dark" size="sm" showGlow showGrid showWave>
        <div className="text-center">
          <LogoWithText className="justify-center" size="lg" variant="light" />
        </div>
      </PageHero>

      <div className="flex-1 px-4 -mt-8 relative z-10 pb-8">
        <div className="w-full max-w-2xl mx-auto">
          <Card padding="lg" variant="elevated" className="shadow-xl">
            <h1 className="text-2xl font-bold text-[var(--text-primary)] mb-6">
              Terms of Service
            </h1>

            <div className="prose prose-sm text-[var(--text-secondary)] space-y-4">
              <p>
                Welcome to Nivo Money. By using our services, you agree to these terms.
              </p>

              <h2 className="text-lg font-semibold text-[var(--text-primary)] mt-6">
                1. Service Description
              </h2>
              <p>
                Nivo Money is a digital wallet and payment platform that allows users to
                store, send, and receive money electronically.
              </p>

              <h2 className="text-lg font-semibold text-[var(--text-primary)] mt-6">
                2. User Responsibilities
              </h2>
              <p>
                Users must provide accurate information during registration and complete
                KYC verification as required by applicable regulations.
              </p>

              <h2 className="text-lg font-semibold text-[var(--text-primary)] mt-6">
                3. Security
              </h2>
              <p>
                Users are responsible for maintaining the confidentiality of their
                account credentials and for all activities under their account.
              </p>

              <h2 className="text-lg font-semibold text-[var(--text-primary)] mt-6">
                4. Contact
              </h2>
              <p>
                For questions about these terms, please contact support@nivomoney.com.
              </p>
            </div>

            <div className="mt-8 flex gap-4">
              <Link to="/login">
                <Button variant="secondary">Back to Login</Button>
              </Link>
              <Link to="/register">
                <Button>Create Account</Button>
              </Link>
            </div>
          </Card>
        </div>
      </div>
    </main>
  );
}

import { Link } from 'react-router-dom';
import { LogoWithText, Card, Button, PageHero } from '@nivo/shared';

export function Privacy() {
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
              Privacy Policy
            </h1>

            <div className="prose prose-sm text-[var(--text-secondary)] space-y-4">
              <p>
                Your privacy is important to us. This policy explains how we collect,
                use, and protect your personal information.
              </p>

              <h2 className="text-lg font-semibold text-[var(--text-primary)] mt-6">
                1. Information We Collect
              </h2>
              <p>
                We collect information you provide during registration, including your
                name, email, phone number, and KYC documents as required by regulations.
              </p>

              <h2 className="text-lg font-semibold text-[var(--text-primary)] mt-6">
                2. How We Use Your Information
              </h2>
              <p>
                We use your information to provide our services, verify your identity,
                process transactions, and comply with legal requirements.
              </p>

              <h2 className="text-lg font-semibold text-[var(--text-primary)] mt-6">
                3. Data Security
              </h2>
              <p>
                We implement industry-standard security measures to protect your data,
                including encryption and secure authentication.
              </p>

              <h2 className="text-lg font-semibold text-[var(--text-primary)] mt-6">
                4. Contact
              </h2>
              <p>
                For privacy-related inquiries, please contact privacy@nivomoney.com.
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

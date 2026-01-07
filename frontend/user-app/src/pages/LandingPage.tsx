import { useNavigate } from 'react-router-dom';

const LandingPage = () => {
  const navigate = useNavigate();

  return (
    <div className="min-h-screen bg-[var(--surface-card)]">
      {/* Hero Section */}
      <section className="relative bg-gradient-to-br from-primary-600 to-primary-800 text-white overflow-hidden">
        {/* Background Pattern */}
        <div className="absolute inset-0 opacity-10">
          <div className="absolute top-0 left-0 w-96 h-96 bg-white rounded-full -translate-x-1/2 -translate-y-1/2"></div>
          <div className="absolute bottom-0 right-0 w-96 h-96 bg-white rounded-full translate-x-1/2 translate-y-1/2"></div>
        </div>

        {/* Navigation */}
        <nav className="relative container mx-auto px-6 py-6">
          <div className="flex justify-between items-center">
            <div className="flex items-center space-x-2">
              <div className="w-10 h-10 bg-white rounded-lg flex items-center justify-center">
                <span className="text-2xl font-bold text-primary-600">N</span>
              </div>
              <span className="text-2xl font-bold">Nivo Money</span>
            </div>
            <div className="flex items-center space-x-4">
              <button
                onClick={() => navigate('/login')}
                className="px-6 py-2 text-white hover:text-primary-100 transition-colors"
              >
                Sign In
              </button>
              <button
                onClick={() => navigate('/register')}
                className="px-6 py-2 bg-white text-primary-600 rounded-lg hover:bg-primary-50 transition-colors font-semibold"
              >
                Get Started Free
              </button>
            </div>
          </div>
        </nav>

        {/* Hero Content */}
        <div className="relative container mx-auto px-6 py-20 lg:py-32">
          <div className="max-w-4xl mx-auto text-center">
            <h1 className="text-5xl lg:text-6xl font-bold mb-6 leading-tight">
              Send Money Instantly,<br />
              <span className="text-primary-200">Securely</span>
            </h1>
            <p className="text-xl lg:text-2xl text-primary-100 mb-8 max-w-2xl mx-auto">
              Fast, free transfers with bank-level security. Your digital wallet for the modern world.
            </p>
            <div className="flex flex-col sm:flex-row gap-4 justify-center">
              <button
                onClick={() => navigate('/register')}
                className="px-8 py-4 bg-white text-primary-600 rounded-lg hover:bg-primary-50 transition-colors font-semibold text-lg shadow-xl"
              >
                Get Started Free →
              </button>
              <button
                onClick={() => navigate('/login')}
                className="px-8 py-4 bg-primary-700 text-white rounded-lg hover:bg-primary-800 transition-colors font-semibold text-lg border-2 border-white/20"
              >
                Sign In
              </button>
            </div>

            {/* Trust Indicators */}
            <div className="mt-12 flex flex-wrap justify-center gap-8 text-primary-100">
              <div className="flex items-center space-x-2">
                <svg className="w-5 h-5" fill="currentColor" viewBox="0 0 20 20">
                  <path fillRule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zm3.707-9.293a1 1 0 00-1.414-1.414L9 10.586 7.707 9.293a1 1 0 00-1.414 1.414l2 2a1 1 0 001.414 0l4-4z" clipRule="evenodd" />
                </svg>
                <span>Bank-Level Security</span>
              </div>
              <div className="flex items-center space-x-2">
                <svg className="w-5 h-5" fill="currentColor" viewBox="0 0 20 20">
                  <path fillRule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zm3.707-9.293a1 1 0 00-1.414-1.414L9 10.586 7.707 9.293a1 1 0 00-1.414 1.414l2 2a1 1 0 001.414 0l4-4z" clipRule="evenodd" />
                </svg>
                <span>Instant Transfers</span>
              </div>
              <div className="flex items-center space-x-2">
                <svg className="w-5 h-5" fill="currentColor" viewBox="0 0 20 20">
                  <path fillRule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zm3.707-9.293a1 1 0 00-1.414-1.414L9 10.586 7.707 9.293a1 1 0 00-1.414 1.414l2 2a1 1 0 001.414 0l4-4z" clipRule="evenodd" />
                </svg>
                <span>Zero Fees</span>
              </div>
            </div>
          </div>
        </div>

        {/* Wave Separator */}
        <div className="relative h-16">
          <svg className="absolute bottom-0 w-full h-16" preserveAspectRatio="none" viewBox="0 0 1440 74" fill="none">
            <path d="M0 74V0C240 49.3333 480 74 720 74C960 74 1200 49.3333 1440 0V74H0Z" fill="white"/>
          </svg>
        </div>
      </section>

      {/* Features Section */}
      <section className="py-20 bg-[var(--surface-card)]">
        <div className="container mx-auto px-6">
          <div className="text-center mb-16">
            <h2 className="text-4xl font-bold text-[var(--text-primary)] mb-4">Why Choose Nivo Money?</h2>
            <p className="text-xl text-[var(--text-secondary)] max-w-2xl mx-auto">
              Everything you need for seamless digital payments, all in one place.
            </p>
          </div>

          <div className="grid md:grid-cols-2 lg:grid-cols-4 gap-8">
            {/* Feature 1 */}
            <div className="text-center p-6 rounded-xl hover:shadow-lg transition-shadow">
              <div className="w-16 h-16 bg-[var(--surface-brand-subtle)] rounded-full flex items-center justify-center mx-auto mb-4">
                <svg className="w-8 h-8 text-[var(--interactive-primary)]" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M13 10V3L4 14h7v7l9-11h-7z" />
                </svg>
              </div>
              <h3 className="text-xl font-bold text-[var(--text-primary)] mb-2">Instant Transfers</h3>
              <p className="text-[var(--text-secondary)]">
                Send money to anyone instantly. No waiting, no delays. Real-time transactions 24/7.
              </p>
            </div>

            {/* Feature 2 */}
            <div className="text-center p-6 rounded-xl hover:shadow-lg transition-shadow">
              <div className="w-16 h-16 bg-[var(--surface-brand-subtle)] rounded-full flex items-center justify-center mx-auto mb-4">
                <svg className="w-8 h-8 text-[var(--interactive-primary)]" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 15v2m-6 4h12a2 2 0 002-2v-6a2 2 0 00-2-2H6a2 2 0 00-2 2v6a2 2 0 002 2zm10-10V7a4 4 0 00-8 0v4h8z" />
                </svg>
              </div>
              <h3 className="text-xl font-bold text-[var(--text-primary)] mb-2">Bank-Level Security</h3>
              <p className="text-[var(--text-secondary)]">
                Your money is protected with enterprise-grade encryption and multi-layer security.
              </p>
            </div>

            {/* Feature 3 */}
            <div className="text-center p-6 rounded-xl hover:shadow-lg transition-shadow">
              <div className="w-16 h-16 bg-[var(--surface-brand-subtle)] rounded-full flex items-center justify-center mx-auto mb-4">
                <svg className="w-8 h-8 text-[var(--interactive-primary)]" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 8c-1.657 0-3 .895-3 2s1.343 2 3 2 3 .895 3 2-1.343 2-3 2m0-8c1.11 0 2.08.402 2.599 1M12 8V7m0 1v8m0 0v1m0-1c-1.11 0-2.08-.402-2.599-1M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
                </svg>
              </div>
              <h3 className="text-xl font-bold text-[var(--text-primary)] mb-2">Zero Fees</h3>
              <p className="text-[var(--text-secondary)]">
                No hidden charges. No transaction fees. Send and receive money absolutely free.
              </p>
            </div>

            {/* Feature 4 */}
            <div className="text-center p-6 rounded-xl hover:shadow-lg transition-shadow">
              <div className="w-16 h-16 bg-[var(--surface-brand-subtle)] rounded-full flex items-center justify-center mx-auto mb-4">
                <svg className="w-8 h-8 text-[var(--interactive-primary)]" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z" />
                </svg>
              </div>
              <h3 className="text-xl font-bold text-[var(--text-primary)] mb-2">24/7 Access</h3>
              <p className="text-[var(--text-secondary)]">
                Access your money anytime, anywhere. Mobile-first design for on-the-go transactions.
              </p>
            </div>
          </div>
        </div>
      </section>

      {/* How It Works Section */}
      <section className="py-20 bg-[var(--surface-page)]">
        <div className="container mx-auto px-6">
          <div className="text-center mb-16">
            <h2 className="text-4xl font-bold text-[var(--text-primary)] mb-4">How It Works</h2>
            <p className="text-xl text-[var(--text-secondary)] max-w-2xl mx-auto">
              Get started in minutes with our simple 4-step process.
            </p>
          </div>

          <div className="max-w-4xl mx-auto">
            <div className="grid md:grid-cols-2 gap-12">
              {/* Step 1 */}
              <div className="flex gap-4">
                <div className="flex-shrink-0">
                  <div className="w-12 h-12 bg-[var(--interactive-primary)] text-white rounded-full flex items-center justify-center font-bold text-lg">
                    1
                  </div>
                </div>
                <div>
                  <h3 className="text-xl font-bold text-[var(--text-primary)] mb-2">Sign Up Free</h3>
                  <p className="text-[var(--text-secondary)]">
                    Create your account in just 2 minutes. All you need is your email and phone number.
                  </p>
                </div>
              </div>

              {/* Step 2 */}
              <div className="flex gap-4">
                <div className="flex-shrink-0">
                  <div className="w-12 h-12 bg-[var(--interactive-primary)] text-white rounded-full flex items-center justify-center font-bold text-lg">
                    2
                  </div>
                </div>
                <div>
                  <h3 className="text-xl font-bold text-[var(--text-primary)] mb-2">Verify Your Identity</h3>
                  <p className="text-[var(--text-secondary)]">
                    Quick KYC verification to ensure security. Upload your documents and get verified.
                  </p>
                </div>
              </div>

              {/* Step 3 */}
              <div className="flex gap-4">
                <div className="flex-shrink-0">
                  <div className="w-12 h-12 bg-[var(--interactive-primary)] text-white rounded-full flex items-center justify-center font-bold text-lg">
                    3
                  </div>
                </div>
                <div>
                  <h3 className="text-xl font-bold text-[var(--text-primary)] mb-2">Add Money</h3>
                  <p className="text-[var(--text-secondary)]">
                    Load your wallet via UPI, bank transfer, or cards. Start with any amount you like.
                  </p>
                </div>
              </div>

              {/* Step 4 */}
              <div className="flex gap-4">
                <div className="flex-shrink-0">
                  <div className="w-12 h-12 bg-[var(--interactive-primary)] text-white rounded-full flex items-center justify-center font-bold text-lg">
                    4
                  </div>
                </div>
                <div>
                  <h3 className="text-xl font-bold text-[var(--text-primary)] mb-2">Send & Receive</h3>
                  <p className="text-[var(--text-secondary)]">
                    Start sending and receiving money instantly. Track all transactions in real-time.
                  </p>
                </div>
              </div>
            </div>
          </div>
        </div>
      </section>

      {/* For Engineers Section */}
      <section className="py-20 bg-[var(--surface-card)] border-t border-[var(--border-subtle)]">
        <div className="container mx-auto px-6">
          <div className="text-center mb-12">
            <span className="inline-block px-4 py-1 bg-[var(--surface-muted)] text-[var(--text-secondary)] rounded-full text-sm font-medium mb-4">
              Portfolio Project
            </span>
            <h2 className="text-4xl font-bold text-[var(--text-primary)] mb-4">Built for Engineers</h2>
            <p className="text-xl text-[var(--text-secondary)] max-w-2xl mx-auto">
              A production-grade microservices architecture demonstrating fintech engineering patterns.
            </p>
          </div>

          {/* Architecture Overview */}
          <div className="max-w-5xl mx-auto mb-12">
            <div className="bg-[var(--surface-page)] rounded-2xl p-8 border border-[var(--border-default)]">
              <div className="grid md:grid-cols-3 gap-6 mb-8">
                <div className="text-center">
                  <div className="text-4xl font-bold text-[var(--interactive-primary)] mb-1">9</div>
                  <div className="text-[var(--text-secondary)]">Microservices</div>
                </div>
                <div className="text-center">
                  <div className="text-4xl font-bold text-[var(--interactive-primary)] mb-1">Go</div>
                  <div className="text-[var(--text-secondary)]">Backend Language</div>
                </div>
                <div className="text-center">
                  <div className="text-4xl font-bold text-[var(--interactive-primary)] mb-1">100%</div>
                  <div className="text-[var(--text-secondary)]">Open Source</div>
                </div>
              </div>

              <div className="grid md:grid-cols-2 gap-8">
                {/* Key Technologies */}
                <div>
                  <h3 className="font-bold text-[var(--text-primary)] mb-4 flex items-center gap-2">
                    <svg className="w-5 h-5 text-[var(--interactive-primary)]" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M10 20l4-16m4 4l4 4-4 4M6 16l-4-4 4-4" />
                    </svg>
                    Key Technologies
                  </h3>
                  <ul className="space-y-2 text-[var(--text-secondary)]">
                    <li className="flex items-center gap-2">
                      <span className="w-2 h-2 bg-[var(--interactive-primary)] rounded-full"></span>
                      Go 1.23 with standard library HTTP
                    </li>
                    <li className="flex items-center gap-2">
                      <span className="w-2 h-2 bg-[var(--interactive-primary)] rounded-full"></span>
                      PostgreSQL with double-entry ledger
                    </li>
                    <li className="flex items-center gap-2">
                      <span className="w-2 h-2 bg-[var(--interactive-primary)] rounded-full"></span>
                      React + TypeScript + Tailwind
                    </li>
                    <li className="flex items-center gap-2">
                      <span className="w-2 h-2 bg-[var(--interactive-primary)] rounded-full"></span>
                      JWT + RBAC authorization
                    </li>
                  </ul>
                </div>

                {/* Architecture Highlights */}
                <div>
                  <h3 className="font-bold text-[var(--text-primary)] mb-4 flex items-center gap-2">
                    <svg className="w-5 h-5 text-[var(--interactive-primary)]" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 11H5m14 0a2 2 0 012 2v6a2 2 0 01-2 2H5a2 2 0 01-2-2v-6a2 2 0 012-2m14 0V9a2 2 0 00-2-2M5 11V9a2 2 0 012-2m0 0V5a2 2 0 012-2h6a2 2 0 012 2v2M7 7h10" />
                    </svg>
                    Architecture Highlights
                  </h3>
                  <ul className="space-y-2 text-[var(--text-secondary)]">
                    <li className="flex items-center gap-2">
                      <span className="w-2 h-2 bg-[var(--interactive-primary)] rounded-full"></span>
                      Domain-driven service boundaries
                    </li>
                    <li className="flex items-center gap-2">
                      <span className="w-2 h-2 bg-[var(--interactive-primary)] rounded-full"></span>
                      Double-entry bookkeeping ledger
                    </li>
                    <li className="flex items-center gap-2">
                      <span className="w-2 h-2 bg-[var(--interactive-primary)] rounded-full"></span>
                      Risk evaluation pipeline
                    </li>
                    <li className="flex items-center gap-2">
                      <span className="w-2 h-2 bg-[var(--interactive-primary)] rounded-full"></span>
                      India-centric (KYC, UPI, INR)
                    </li>
                  </ul>
                </div>
              </div>
            </div>
          </div>

          {/* Links */}
          <div className="flex flex-wrap justify-center gap-4">
            <a
              href="https://github.com/vnykmshr/nivo"
              target="_blank"
              rel="noopener noreferrer"
              className="px-6 py-3 bg-[var(--text-primary)] text-[var(--surface-card)] rounded-lg hover:opacity-90 transition-opacity font-medium inline-flex items-center gap-2"
            >
              <svg className="w-5 h-5" fill="currentColor" viewBox="0 0 24 24">
                <path fillRule="evenodd" d="M12 2C6.477 2 2 6.484 2 12.017c0 4.425 2.865 8.18 6.839 9.504.5.092.682-.217.682-.483 0-.237-.008-.868-.013-1.703-2.782.605-3.369-1.343-3.369-1.343-.454-1.158-1.11-1.466-1.11-1.466-.908-.62.069-.608.069-.608 1.003.07 1.531 1.032 1.531 1.032.892 1.53 2.341 1.088 2.91.832.092-.647.35-1.088.636-1.338-2.22-.253-4.555-1.113-4.555-4.951 0-1.093.39-1.988 1.029-2.688-.103-.253-.446-1.272.098-2.65 0 0 .84-.27 2.75 1.026A9.564 9.564 0 0112 6.844c.85.004 1.705.115 2.504.337 1.909-1.296 2.747-1.027 2.747-1.027.546 1.379.202 2.398.1 2.651.64.7 1.028 1.595 1.028 2.688 0 3.848-2.339 4.695-4.566 4.943.359.309.678.92.678 1.855 0 1.338-.012 2.419-.012 2.747 0 .268.18.58.688.482A10.019 10.019 0 0022 12.017C22 6.484 17.522 2 12 2z" clipRule="evenodd" />
              </svg>
              View on GitHub
            </a>
            <a
              href="https://docs.nivomoney.com"
              target="_blank"
              rel="noopener noreferrer"
              className="px-6 py-3 bg-[var(--interactive-primary)] text-white rounded-lg hover:opacity-90 transition-opacity font-medium inline-flex items-center gap-2"
            >
              <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 6.253v13m0-13C10.832 5.477 9.246 5 7.5 5S4.168 5.477 3 6.253v13C4.168 18.477 5.754 18 7.5 18s3.332.477 4.5 1.253m0-13C13.168 5.477 14.754 5 16.5 5c1.747 0 3.332.477 4.5 1.253v13C19.832 18.477 18.247 18 16.5 18c-1.746 0-3.332.477-4.5 1.253" />
              </svg>
              Documentation
            </a>
            <a
              href="https://docs.nivomoney.com/adr"
              target="_blank"
              rel="noopener noreferrer"
              className="px-6 py-3 border-2 border-[var(--border-default)] text-[var(--text-primary)] rounded-lg hover:border-[var(--border-strong)] transition-colors font-medium inline-flex items-center gap-2"
            >
              <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
              </svg>
              Architecture Decisions
            </a>
          </div>

          {/* Demo Credentials */}
          <div className="mt-12 max-w-md mx-auto">
            <div className="bg-[var(--surface-page)] rounded-xl p-6 border border-[var(--border-default)] text-center">
              <h3 className="font-bold text-[var(--text-primary)] mb-2">Try the Demo</h3>
              <p className="text-sm text-[var(--text-secondary)] mb-4">
                Pre-seeded accounts with test data
              </p>
              <div className="bg-[var(--surface-card)] rounded-lg p-4 font-mono text-sm text-left border border-[var(--border-subtle)]">
                <div className="flex justify-between mb-1">
                  <span className="text-[var(--text-muted)]">Email:</span>
                  <span className="text-[var(--text-primary)]">raj.kumar@gmail.com</span>
                </div>
                <div className="flex justify-between">
                  <span className="text-[var(--text-muted)]">Password:</span>
                  <span className="text-[var(--text-primary)]">raj123</span>
                </div>
              </div>
              <p className="text-xs text-[var(--text-muted)] mt-3">
                Balance: ₹50,000 • KYC Verified
              </p>
            </div>
          </div>
        </div>
      </section>

      {/* CTA Section */}
      <section className="py-20 bg-[var(--interactive-primary)] text-white">
        <div className="container mx-auto px-6 text-center">
          <h2 className="text-4xl font-bold mb-4">Ready to Get Started?</h2>
          <p className="text-xl text-white/80 mb-8 max-w-2xl mx-auto">
            Join thousands of users who trust Nivo Money for their digital transactions.
          </p>
          <button
            onClick={() => navigate('/register')}
            className="px-8 py-4 bg-white text-[var(--interactive-primary)] rounded-lg hover:bg-white/90 transition-colors font-semibold text-lg shadow-xl inline-flex items-center gap-2"
          >
            Create Free Account
            <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M13 7l5 5m0 0l-5 5m5-5H6" />
            </svg>
          </button>
          <p className="mt-4 text-white/60 text-sm">No credit card required • Free forever</p>
        </div>
      </section>

      {/* Footer - Intentionally dark design, fixed colors */}
      <footer className="bg-gray-900 text-gray-300 py-12">
        <div className="container mx-auto px-6">
          <div className="grid md:grid-cols-4 gap-8 mb-8">
            {/* Company */}
            <div>
              <div className="flex items-center space-x-2 mb-4">
                <div className="w-8 h-8 bg-[var(--interactive-primary)] rounded-lg flex items-center justify-center">
                  <span className="text-lg font-bold text-white">N</span>
                </div>
                <span className="text-xl font-bold text-white">Nivo Money</span>
              </div>
              <p className="text-sm text-gray-400">
                Fast, secure digital payments for everyone.
              </p>
            </div>

            {/* Product */}
            <div>
              <h4 className="font-semibold text-white mb-4">Product</h4>
              <ul className="space-y-2 text-sm">
                <li><button onClick={() => window.scrollTo({ top: 0, behavior: 'smooth' })} className="hover:text-white transition-colors text-left">Features</button></li>
                <li><button onClick={() => window.scrollTo({ top: 0, behavior: 'smooth' })} className="hover:text-white transition-colors text-left">Security</button></li>
                <li><button onClick={() => window.scrollTo({ top: 0, behavior: 'smooth' })} className="hover:text-white transition-colors text-left">Pricing</button></li>
              </ul>
            </div>

            {/* Company */}
            <div>
              <h4 className="font-semibold text-white mb-4">Company</h4>
              <ul className="space-y-2 text-sm">
                <li><button onClick={() => window.scrollTo({ top: 0, behavior: 'smooth' })} className="hover:text-white transition-colors text-left">About Us</button></li>
                <li><button onClick={() => window.scrollTo({ top: 0, behavior: 'smooth' })} className="hover:text-white transition-colors text-left">Contact</button></li>
                <li><button onClick={() => window.scrollTo({ top: 0, behavior: 'smooth' })} className="hover:text-white transition-colors text-left">Careers</button></li>
              </ul>
            </div>

            {/* Legal */}
            <div>
              <h4 className="font-semibold text-white mb-4">Legal</h4>
              <ul className="space-y-2 text-sm">
                <li><button onClick={() => window.scrollTo({ top: 0, behavior: 'smooth' })} className="hover:text-white transition-colors text-left">Privacy Policy</button></li>
                <li><button onClick={() => window.scrollTo({ top: 0, behavior: 'smooth' })} className="hover:text-white transition-colors text-left">Terms of Service</button></li>
                <li><button onClick={() => window.scrollTo({ top: 0, behavior: 'smooth' })} className="hover:text-white transition-colors text-left">Compliance</button></li>
              </ul>
            </div>
          </div>

          <div className="border-t border-gray-800 pt-8 text-center text-sm text-gray-400">
            <p>&copy; 2025 Nivo Money by Nivo. All rights reserved.</p>
          </div>
        </div>
      </footer>
    </div>
  );
};

export default LandingPage;

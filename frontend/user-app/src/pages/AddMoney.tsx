import { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import { useWalletStore } from '../stores/walletStore';
import { api } from '../lib/api';
import { formatCurrency, toPaise } from '../lib/utils';
import type { UPIDepositResponse } from '../types';
import {
  Alert,
  Button,
  Card,
  FormField,
  Input,
  Logo,
} from '../../../shared/components';

type Step = 'input' | 'payment' | 'success';

export function AddMoney() {
  const navigate = useNavigate();
  const { wallets, fetchWallets } = useWalletStore();

  // Form state
  const [step, setStep] = useState<Step>('input');
  const [walletId, setWalletId] = useState('');
  const [amount, setAmount] = useState('');

  // Payment state
  const [depositResponse, setDepositResponse] = useState<UPIDepositResponse | null>(null);
  const [isInitiating, setIsInitiating] = useState(false);
  const [isCompleting, setIsCompleting] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [errors, setErrors] = useState<Record<string, string>>({});

  useEffect(() => {
    if (wallets.length === 0) {
      fetchWallets().catch(err => console.error('Failed to fetch wallets:', err));
    }
  }, [wallets.length, fetchWallets]);

  // Auto-select default active wallet (prioritize 'default' type)
  useEffect(() => {
    if (!walletId && wallets.length > 0) {
      const defaultWallet = wallets.find(w => w.type === 'default' && w.status === 'active');
      const activeWallet = wallets.find(w => w.status === 'active');
      const selected = defaultWallet || activeWallet;
      if (selected) {
        setWalletId(selected.id);
      }
    }
  }, [wallets, walletId]);

  const selectedWallet = wallets.find(w => w.id === walletId);

  // Validate form
  const validateForm = (): boolean => {
    const newErrors: Record<string, string> = {};

    if (!walletId) {
      newErrors.walletId = 'Please select a wallet';
    }

    if (!amount) {
      newErrors.amount = 'Please enter an amount';
    } else {
      const amountNum = parseFloat(amount);
      if (isNaN(amountNum) || amountNum <= 0) {
        newErrors.amount = 'Amount must be greater than 0';
      } else if (amountNum < 1) {
        newErrors.amount = 'Minimum amount is ₹1';
      } else if (amountNum > 100000) {
        newErrors.amount = 'Maximum amount is ₹100,000';
      }
    }

    setErrors(newErrors);
    return Object.keys(newErrors).length === 0;
  };

  const handleInitiateDeposit = async () => {
    if (!validateForm()) return;

    setIsInitiating(true);
    setError(null);

    try {
      const response = await api.initiateUPIDeposit({
        wallet_id: walletId,
        amount: toPaise(parseFloat(amount)),
        currency: 'INR',
        description: 'Add money via UPI',
      });

      setDepositResponse(response);
      setStep('payment');
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to initiate deposit');
    } finally {
      setIsInitiating(false);
    }
  };

  const handleSimulatePayment = async () => {
    if (!depositResponse) return;

    setIsCompleting(true);
    setError(null);

    try {
      // Generate a mock UPI transaction ID
      const upiTxId = `UPI${Date.now()}${Math.random().toString(36).substr(2, 9)}`;

      await api.completeUPIDeposit({
        transaction_id: depositResponse.transaction.id,
        upi_transaction_id: upiTxId,
        status: 'success',
      });

      // Refetch wallets to show updated balance
      await fetchWallets();

      // Small delay to ensure backend has processed the transaction
      await new Promise(resolve => setTimeout(resolve, 500));

      setStep('success');
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to complete deposit');
    } finally {
      setIsCompleting(false);
    }
  };

  const handleAddMore = () => {
    setStep('input');
    setAmount('');
    setDepositResponse(null);
    setError(null);
    setErrors({});
  };

  // Input Step
  if (step === 'input') {
    return (
      <div className="min-h-screen bg-[var(--surface-page)]">
        <nav className="bg-[var(--surface-card)] shadow-sm border-b border-[var(--border-subtle)]">
          <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
            <div className="flex justify-between h-16 items-center">
              <Logo className="text-xl font-bold" />
              <Button variant="secondary" onClick={() => navigate('/dashboard')}>
                Back to Dashboard
              </Button>
            </div>
          </div>
        </nav>

        <main className="max-w-2xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
          <Card padding="lg">
            <h2 className="text-2xl font-bold text-[var(--text-primary)] mb-6">Add Money</h2>

            {error && (
              <Alert variant="error" className="mb-4" onDismiss={() => setError(null)}>
                {error}
              </Alert>
            )}

            <div className="space-y-6">
              {/* Wallet Selection */}
              <FormField
                label="To Wallet"
                htmlFor="wallet"
                error={errors.walletId}
                hint={selectedWallet ? `Current Balance: ${formatCurrency(selectedWallet.available_balance)}` : undefined}
              >
                <select
                  id="wallet"
                  value={walletId}
                  onChange={e => {
                    setWalletId(e.target.value);
                    setErrors({ ...errors, walletId: '' });
                  }}
                  className="w-full h-10 px-3 pr-10 text-sm appearance-none rounded-[var(--radius-input)] border border-[var(--input-border)] bg-[var(--input-bg)] text-[var(--input-text)] focus:border-[var(--input-border-focus)] focus:outline-none focus:[box-shadow:var(--focus-ring)] disabled:opacity-50 disabled:cursor-not-allowed"
                >
                  <option value="">Select a wallet</option>
                  {wallets
                    .filter(w => w.status === 'active')
                    .map(wallet => (
                      <option key={wallet.id} value={wallet.id}>
                        {wallet.currency} Wallet - {formatCurrency(wallet.available_balance)}
                      </option>
                    ))}
                </select>
              </FormField>

              {/* Amount */}
              <FormField
                label="Amount (₹)"
                htmlFor="amount"
                error={errors.amount}
                hint="Min: ₹1 | Max: ₹100,000"
              >
                <Input
                  type="number"
                  id="amount"
                  value={amount}
                  onChange={e => {
                    setAmount(e.target.value);
                    setErrors({ ...errors, amount: '' });
                  }}
                  placeholder="0.00"
                  step="0.01"
                  min="1"
                  max="100000"
                  error={!!errors.amount}
                />
              </FormField>

              {/* Quick Amount Buttons */}
              <div>
                <label className="block text-sm font-medium text-[var(--text-primary)] mb-2">
                  Quick Select
                </label>
                <div className="grid grid-cols-4 gap-2">
                  {[100, 500, 1000, 5000].map(amt => (
                    <Button
                      key={amt}
                      type="button"
                      variant="secondary"
                      size="sm"
                      onClick={() => setAmount(amt.toString())}
                    >
                      ₹{amt}
                    </Button>
                  ))}
                </div>
              </div>

              {/* Continue Button */}
              <Button
                type="button"
                onClick={handleInitiateDeposit}
                disabled={!walletId || !amount}
                loading={isInitiating}
                className="w-full"
              >
                Continue
              </Button>
            </div>
          </Card>
        </main>
      </div>
    );
  }

  // Payment Step
  if (step === 'payment' && depositResponse) {
    return (
      <div className="min-h-screen bg-[var(--surface-page)]">
        <nav className="bg-[var(--surface-card)] shadow-sm border-b border-[var(--border-subtle)]">
          <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
            <div className="flex justify-between h-16 items-center">
              <Logo className="text-xl font-bold" />
            </div>
          </div>
        </nav>

        <main className="max-w-2xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
          <Card padding="lg">
            <h2 className="text-2xl font-bold text-[var(--text-primary)] mb-6">Complete Payment</h2>

            {error && (
              <Alert variant="error" className="mb-4" onDismiss={() => setError(null)}>
                {error}
              </Alert>
            )}

            <div className="space-y-6">
              {/* Amount Display */}
              <div className="text-center bg-[var(--surface-brand-subtle)] rounded-lg p-6">
                <p className="text-sm text-[var(--text-secondary)] mb-2">Amount to Pay</p>
                <p className="text-4xl font-bold text-[var(--interactive-primary)]">
                  {formatCurrency(depositResponse.transaction.amount)}
                </p>
              </div>

              {/* Virtual UPI ID */}
              <div className="bg-[var(--surface-muted)] rounded-lg p-4">
                <label className="block text-sm font-medium text-[var(--text-primary)] mb-2">
                  Pay to this UPI ID
                </label>
                <div className="flex items-center gap-2">
                  <Input
                    type="text"
                    value={depositResponse.virtual_upi_id}
                    readOnly
                    className="flex-1 font-mono text-sm"
                  />
                  <Button
                    type="button"
                    variant="secondary"
                    onClick={() => {
                      navigator.clipboard.writeText(depositResponse.virtual_upi_id);
                    }}
                  >
                    Copy
                  </Button>
                </div>
              </div>

              {/* QR Code (Mock) */}
              <div className="text-center">
                <p className="text-sm text-[var(--text-secondary)] mb-3">Or scan this QR code</p>
                <div className="inline-block p-4 bg-[var(--surface-card)] border-2 border-[var(--border-default)] rounded-lg">
                  <img
                    src={depositResponse.qr_code}
                    alt="UPI QR Code"
                    className="w-48 h-48 mx-auto"
                  />
                </div>
              </div>

              {/* Instructions */}
              <Alert variant="info">
                <h3 className="font-semibold mb-2">Payment Instructions</h3>
                <ol className="list-decimal list-inside space-y-1 text-sm">
                  {depositResponse.instructions.map((instruction, idx) => (
                    <li key={idx}>{instruction}</li>
                  ))}
                </ol>
              </Alert>

              {/* Expiry Warning */}
              <Alert variant="warning">
                <p className="text-sm">
                  ⏰ This payment link expires at{' '}
                  <span className="font-semibold">
                    {new Date(depositResponse.expires_at).toLocaleTimeString()}
                  </span>
                </p>
              </Alert>

              {/* Simulate Payment Button (for demo) */}
              <div className="border-t border-[var(--border-default)] pt-6">
                <Button
                  type="button"
                  onClick={handleSimulatePayment}
                  loading={isCompleting}
                  className="w-full"
                >
                  ✨ Simulate Payment (Demo)
                </Button>
                <p className="text-xs text-[var(--text-muted)] text-center mt-2">
                  In production, this would happen automatically when you complete payment in your UPI app
                </p>
              </div>

              <Button
                type="button"
                variant="secondary"
                onClick={() => navigate('/dashboard')}
                className="w-full"
              >
                Cancel
              </Button>
            </div>
          </Card>
        </main>
      </div>
    );
  }

  // Success Step
  if (step === 'success' && depositResponse) {
    return (
      <div className="min-h-screen bg-[var(--surface-page)]">
        <nav className="bg-[var(--surface-card)] shadow-sm border-b border-[var(--border-subtle)]">
          <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
            <div className="flex justify-between h-16 items-center">
              <Logo className="text-xl font-bold" />
            </div>
          </div>
        </nav>

        <main className="max-w-2xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
          <Card padding="lg">
            <div className="text-center">
              {/* Success Icon */}
              <div className="w-16 h-16 bg-[var(--surface-success)] rounded-full flex items-center justify-center mx-auto mb-4">
                <svg className="w-8 h-8 text-[var(--text-success)]" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M5 13l4 4L19 7" />
                </svg>
              </div>

              <h2 className="text-2xl font-bold text-[var(--text-primary)] mb-2">Money Added Successfully!</h2>
              <p className="text-[var(--text-secondary)] mb-6">
                Your wallet has been credited with{' '}
                <span className="font-semibold text-[var(--interactive-primary)]">
                  {formatCurrency(depositResponse.transaction.amount)}
                </span>
              </p>

              {/* Transaction Details */}
              <div className="bg-[var(--surface-muted)] rounded-lg p-4 mb-6 text-left">
                <h3 className="font-semibold text-[var(--text-primary)] mb-3">Transaction Details</h3>
                <div className="space-y-2 text-sm">
                  <div className="flex justify-between">
                    <span className="text-[var(--text-secondary)]">Transaction ID:</span>
                    <span className="font-mono text-[var(--text-primary)]">{depositResponse.transaction.id.substring(0, 12)}...</span>
                  </div>
                  <div className="flex justify-between">
                    <span className="text-[var(--text-secondary)]">UPI ID:</span>
                    <span className="font-mono text-[var(--text-primary)]">{depositResponse.virtual_upi_id}</span>
                  </div>
                  <div className="flex justify-between">
                    <span className="text-[var(--text-secondary)]">Status:</span>
                    <span className="text-[var(--text-success)] font-semibold">✓ Completed</span>
                  </div>
                </div>
              </div>

              {/* Action Buttons */}
              <div className="space-y-3">
                <Button
                  type="button"
                  onClick={handleAddMore}
                  className="w-full"
                >
                  Add More Money
                </Button>
                <Button
                  type="button"
                  variant="secondary"
                  onClick={() => navigate('/dashboard')}
                  className="w-full"
                >
                  Back to Dashboard
                </Button>
              </div>
            </div>
          </Card>
        </main>
      </div>
    );
  }

  return null;
}

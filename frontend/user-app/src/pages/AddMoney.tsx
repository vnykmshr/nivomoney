import { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import { useWalletStore } from '../stores/walletStore';
import { api } from '../lib/api';
import { formatCurrency, toPaise } from '../lib/utils';
import type { UPIDepositResponse } from '@nivo/shared';
import { AppLayout } from '../components';
import {
  Alert,
  Button,
  Card,
  FormField,
  Input,
  StepIndicator,
  AmountDisplay,
  SuccessState,
  TrustBadge,
} from '@nivo/shared';

type Step = 'input' | 'payment' | 'success';

const STEPS = ['Enter Amount', 'Pay via UPI', 'Done'];

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

  const currentStepIndex = step === 'input' ? 0 : step === 'payment' ? 1 : 2;

  useEffect(() => {
    if (wallets.length === 0) {
      fetchWallets().catch(() => {/* Wallet fetch error handled by store */});
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
      <AppLayout title="Add Money" showBack>
        <div className="max-w-md mx-auto px-4 py-6">
          {/* Step Indicator */}
          <div className="mb-6">
            <StepIndicator steps={STEPS} currentStep={currentStepIndex} variant="numbered" />
          </div>

          {error && (
            <Alert variant="error" onDismiss={() => setError(null)} className="mb-4">
              {error}
            </Alert>
          )}

          <Card padding="lg">
            <div className="space-y-6">
              {/* Wallet Selection */}
              <FormField
                label="To Wallet"
                htmlFor="wallet"
                error={errors.walletId}
                hint={selectedWallet ? `Current Balance: ${formatCurrency(selectedWallet.available_balance)}` : undefined}
                required
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
                required
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
                size="lg"
              >
                Continue
              </Button>
            </div>
          </Card>

          {/* Trust Badge */}
          <div className="flex justify-center mt-6">
            <TrustBadge variant="security" size="sm" theme="light" />
          </div>
        </div>
      </AppLayout>
    );
  }

  // Payment Step
  if (step === 'payment' && depositResponse) {
    return (
      <AppLayout title="Complete Payment">
        <div className="max-w-md mx-auto px-4 py-6">
          {/* Step Indicator */}
          <div className="mb-6">
            <StepIndicator steps={STEPS} currentStep={currentStepIndex} variant="numbered" />
          </div>

          {error && (
            <Alert variant="error" onDismiss={() => setError(null)} className="mb-4">
              {error}
            </Alert>
          )}

          <Card padding="lg">
            <div className="space-y-6">
              {/* Amount Display */}
              <AmountDisplay
                amount={depositResponse.transaction.amount}
                label="Amount to Pay"
                size="xl"
                variant="highlight"
                showBackground
              />

              {/* Virtual UPI ID */}
              <div className="bg-[var(--surface-page)] rounded-lg p-4">
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

              {/* QR Code */}
              <div className="text-center">
                <p className="text-sm text-[var(--text-secondary)] mb-3">Or scan this QR code</p>
                <div className="inline-block p-4 bg-[var(--surface-card)] border-2 border-[var(--border-default)] rounded-xl">
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
                  This payment link expires at{' '}
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
                  size="lg"
                >
                  Simulate Payment (Demo)
                </Button>
                <p className="text-xs text-[var(--text-muted)] text-center mt-2">
                  In production, this happens automatically when you complete payment in your UPI app
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

          {/* Trust Badge */}
          <div className="flex justify-center mt-6">
            <TrustBadge variant="encrypted" size="sm" theme="light" />
          </div>
        </div>
      </AppLayout>
    );
  }

  // Success Step
  if (step === 'success' && depositResponse) {
    const successDetails = [
      { label: 'Transaction ID', value: `${depositResponse.transaction.id.substring(0, 12)}...` },
      { label: 'UPI ID', value: depositResponse.virtual_upi_id },
      { label: 'Status', value: 'Completed' },
    ];

    return (
      <AppLayout title="Success">
        <div className="max-w-md mx-auto px-4 py-6">
          {/* Step Indicator */}
          <div className="mb-6">
            <StepIndicator steps={STEPS} currentStep={currentStepIndex} variant="numbered" />
          </div>

          <Card padding="lg">
            <SuccessState
              title="Money Added Successfully!"
              message={`${formatCurrency(depositResponse.transaction.amount)} has been added to your wallet`}
              icon="money"
              showAnimation
              details={successDetails}
              primaryAction={{
                label: 'Back to Dashboard',
                onClick: () => navigate('/dashboard'),
              }}
              secondaryAction={{
                label: 'Add More Money',
                onClick: handleAddMore,
              }}
            />
          </Card>
        </div>
      </AppLayout>
    );
  }

  return null;
}

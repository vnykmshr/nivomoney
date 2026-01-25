import { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import { useWalletStore } from '../stores/walletStore';
import { api } from '../lib/api';
import { formatCurrency, toPaise } from '../lib/utils';
import { AppLayout } from '../components';
import {
  Card,
  Button,
  Alert,
  FormField,
  Input,
  StepIndicator,
  SuccessState,
  TrustBadge,
} from '@nivo/shared';

type Step = 'form' | 'success';

const STEPS = ['Enter Details', 'Done'];

export function Deposit() {
  const navigate = useNavigate();
  const { wallets, fetchWallets } = useWalletStore();

  const [step, setStep] = useState<Step>('form');
  const [walletId, setWalletId] = useState('');
  const [amount, setAmount] = useState('');
  const [paymentMethod, setPaymentMethod] = useState('');
  const [reference, setReference] = useState('');
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [errors, setErrors] = useState<Record<string, string>>({});
  const [depositedAmount, setDepositedAmount] = useState(0);

  const currentStepIndex = step === 'form' ? 0 : 1;

  useEffect(() => {
    if (wallets.length === 0) {
      fetchWallets().catch(() => {});
    }
  }, [wallets.length, fetchWallets]);

  // Auto-select default active wallet
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

  const validate = (): boolean => {
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
      }
    }

    if (!paymentMethod) {
      newErrors.paymentMethod = 'Please select a payment method';
    }

    setErrors(newErrors);
    return Object.keys(newErrors).length === 0;
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError(null);

    if (!validate()) return;

    setIsLoading(true);

    try {
      const amountPaise = toPaise(parseFloat(amount));
      setDepositedAmount(amountPaise);

      await api.createDeposit({
        wallet_id: walletId,
        amount_paise: amountPaise,
        description: `Deposit via ${paymentMethod}`,
        reference: reference || undefined,
      });

      await fetchWallets();
      setStep('success');
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to process deposit');
    } finally {
      setIsLoading(false);
    }
  };

  const handleNewDeposit = () => {
    setStep('form');
    setAmount('');
    setPaymentMethod('');
    setReference('');
    setError(null);
    setErrors({});
  };

  const selectedWallet = wallets.find(w => w.id === walletId);

  const paymentMethodLabels: Record<string, string> = {
    bank_transfer: 'Bank Transfer',
    upi: 'UPI',
    card: 'Card',
    net_banking: 'Net Banking',
  };

  // Form Step
  if (step === 'form') {
    return (
      <AppLayout title="Deposit Money" showBack>
        <div className="max-w-md mx-auto px-4 py-6">
          {/* Step Indicator */}
          <div className="mb-6">
            <StepIndicator steps={STEPS} currentStep={currentStepIndex} variant="numbered" />
          </div>

          {error && (
            <Alert variant="error" className="mb-4" onDismiss={() => setError(null)}>
              {error}
            </Alert>
          )}

          <Card padding="lg">
            <form onSubmit={handleSubmit} className="space-y-5">
              {/* Wallet Selection */}
              <FormField
                label="To Wallet"
                htmlFor="wallet"
                error={errors.walletId}
                hint={selectedWallet ? `Current Balance: ${formatCurrency(selectedWallet.balance)}` : undefined}
                required
              >
                <select
                  id="wallet"
                  value={walletId}
                  onChange={e => setWalletId(e.target.value)}
                  className="w-full h-10 px-3 pr-10 text-sm appearance-none rounded-[var(--radius-input)] border border-[var(--input-border)] bg-[var(--input-bg)] text-[var(--input-text)] focus:border-[var(--input-border-focus)] focus:outline-none focus:[box-shadow:var(--focus-ring)] disabled:opacity-50 disabled:cursor-not-allowed"
                  disabled={isLoading}
                >
                  <option value="">Select a wallet</option>
                  {wallets
                    .filter(w => w.status === 'active')
                    .map(wallet => (
                      <option key={wallet.id} value={wallet.id}>
                        {wallet.type.toUpperCase()} - {formatCurrency(wallet.balance)}
                      </option>
                    ))}
                </select>
              </FormField>

              {/* Amount */}
              <FormField
                label="Amount (₹)"
                htmlFor="amount"
                error={errors.amount}
                required
              >
                <Input
                  type="number"
                  id="amount"
                  value={amount}
                  onChange={e => setAmount(e.target.value)}
                  placeholder="0.00"
                  disabled={isLoading}
                  error={!!errors.amount}
                />
              </FormField>

              {/* Payment Method */}
              <FormField
                label="Payment Method"
                htmlFor="paymentMethod"
                error={errors.paymentMethod}
                required
              >
                <select
                  id="paymentMethod"
                  value={paymentMethod}
                  onChange={e => setPaymentMethod(e.target.value)}
                  className="w-full h-10 px-3 pr-10 text-sm appearance-none rounded-[var(--radius-input)] border border-[var(--input-border)] bg-[var(--input-bg)] text-[var(--input-text)] focus:border-[var(--input-border-focus)] focus:outline-none focus:[box-shadow:var(--focus-ring)] disabled:opacity-50 disabled:cursor-not-allowed"
                  disabled={isLoading}
                >
                  <option value="">Select payment method</option>
                  <option value="bank_transfer">Bank Transfer</option>
                  <option value="upi">UPI</option>
                  <option value="card">Card</option>
                  <option value="net_banking">Net Banking</option>
                </select>
              </FormField>

              {/* Reference (Optional) */}
              <FormField
                label="Reference Number (Optional)"
                htmlFor="reference"
              >
                <Input
                  type="text"
                  id="reference"
                  value={reference}
                  onChange={e => setReference(e.target.value)}
                  placeholder="Transaction reference"
                  disabled={isLoading}
                />
              </FormField>

              {/* Submit Button */}
              <Button
                type="submit"
                className="w-full"
                size="lg"
                loading={isLoading}
              >
                Deposit Money
              </Button>
            </form>
          </Card>

          {/* Trust Badge */}
          <div className="flex justify-center mt-6">
            <TrustBadge variant="security" size="sm" theme="light" />
          </div>
        </div>
      </AppLayout>
    );
  }

  // Success Step
  const successDetails = [
    { label: 'Amount', value: formatCurrency(depositedAmount) },
    { label: 'Payment Method', value: paymentMethodLabels[paymentMethod] || paymentMethod },
    ...(reference ? [{ label: 'Reference', value: reference }] : []),
    { label: 'Status', value: 'Processing' },
  ];

  return (
    <AppLayout title="Deposit Initiated">
      <div className="max-w-md mx-auto px-4 py-6">
        {/* Step Indicator */}
        <div className="mb-6">
          <StepIndicator steps={STEPS} currentStep={currentStepIndex} variant="numbered" />
        </div>

        <Card padding="lg">
          <SuccessState
            title="Deposit Initiated!"
            message={`${formatCurrency(depositedAmount)} will be added to your wallet`}
            icon="money"
            showAnimation
            details={successDetails}
            primaryAction={{
              label: 'Back to Dashboard',
              onClick: () => navigate('/dashboard'),
            }}
            secondaryAction={{
              label: 'Deposit Again',
              onClick: handleNewDeposit,
            }}
          />
        </Card>
      </div>
    </AppLayout>
  );
}

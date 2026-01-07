import { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import { useWalletStore } from '../stores/walletStore';
import { api } from '../lib/api';
import { formatCurrency, toPaise } from '../lib/utils';
import { AppLayout } from '../components';
import {
  Card,
  CardTitle,
  Button,
  Alert,
  FormField,
  Input,
} from '../../../shared/components';

export function Deposit() {
  const navigate = useNavigate();
  const { wallets, fetchWallets } = useWalletStore();
  const [walletId, setWalletId] = useState('');
  const [amount, setAmount] = useState('');
  const [paymentMethod, setPaymentMethod] = useState('');
  const [reference, setReference] = useState('');
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [success, setSuccess] = useState(false);
  const [errors, setErrors] = useState<Record<string, string>>({});

  useEffect(() => {
    if (wallets.length === 0) {
      fetchWallets().catch(() => {
        // Error handled in store
      });
    }
  }, [wallets.length, fetchWallets]);

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
    setSuccess(false);

    if (!validate()) return;

    setIsLoading(true);

    try {
      await api.createDeposit({
        wallet_id: walletId,
        amount_paise: toPaise(parseFloat(amount)),
        description: `Deposit via ${paymentMethod}`,
        reference: reference || undefined,
      });

      setSuccess(true);
      setWalletId('');
      setAmount('');
      setPaymentMethod('');
      setReference('');

      // Refetch wallets to update balance
      await fetchWallets();

      // Navigate back to dashboard after 2 seconds
      setTimeout(() => {
        navigate('/dashboard');
      }, 2000);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to process deposit');
      setIsLoading(false);
    }
  };

  const selectedWallet = wallets.find(w => w.id === walletId);

  return (
    <AppLayout title="Deposit Money">
      <div className="max-w-2xl mx-auto">
        <Card>
          <CardTitle className="mb-6">Deposit Money</CardTitle>

          {error && (
            <Alert variant="error" className="mb-4" onDismiss={() => setError(null)}>
              {error}
            </Alert>
          )}

          {success && (
            <Alert variant="success" className="mb-4">
              Deposit initiated successfully! Redirecting to dashboard...
            </Alert>
          )}

          <form onSubmit={handleSubmit} className="space-y-6">
            {/* Wallet Selection */}
            <FormField
              label="To Wallet"
              htmlFor="wallet"
              error={errors.walletId}
            >
              <select
                id="wallet"
                value={walletId}
                onChange={e => setWalletId(e.target.value)}
                className="w-full px-4 py-3 rounded-lg border bg-[var(--surface-input)] border-[var(--border-default)] text-[var(--text-primary)] focus:outline-none focus:ring-2 focus:ring-[var(--interactive-primary)] focus:border-transparent"
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
            {selectedWallet && (
              <p className="text-sm text-[var(--text-muted)] -mt-4">
                Current balance: {formatCurrency(selectedWallet.balance)}
              </p>
            )}

            {/* Amount */}
            <FormField
              label="Amount (₹)"
              htmlFor="amount"
              error={errors.amount}
            >
              <Input
                type="number"
                id="amount"
                value={amount}
                onChange={e => setAmount(e.target.value)}
                placeholder="0.00"
                disabled={isLoading}
              />
            </FormField>

            {/* Payment Method */}
            <FormField
              label="Payment Method"
              htmlFor="paymentMethod"
              error={errors.paymentMethod}
            >
              <select
                id="paymentMethod"
                value={paymentMethod}
                onChange={e => setPaymentMethod(e.target.value)}
                className="w-full px-4 py-3 rounded-lg border bg-[var(--surface-input)] border-[var(--border-default)] text-[var(--text-primary)] focus:outline-none focus:ring-2 focus:ring-[var(--interactive-primary)] focus:border-transparent"
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
              disabled={isLoading}
              loading={isLoading}
            >
              Deposit Money
            </Button>
          </form>
        </Card>
      </div>
    </AppLayout>
  );
}

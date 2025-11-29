import { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import { useWalletStore } from '../stores/walletStore';
import { api } from '../lib/api';
import { formatCurrency, toPaise } from '../lib/utils';

export function Withdraw() {
  const navigate = useNavigate();
  const { wallets, fetchWallets } = useWalletStore();
  const [walletId, setWalletId] = useState('');
  const [amount, setAmount] = useState('');
  const [bankAccount, setBankAccount] = useState('');
  const [ifscCode, setIfscCode] = useState('');
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [success, setSuccess] = useState(false);
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
      } else {
        const sourceWallet = wallets.find(w => w.id === walletId);
        if (sourceWallet && amountNum * 100 > sourceWallet.available_balance) {
          newErrors.amount = `Insufficient balance. Available: ${formatCurrency(sourceWallet.available_balance)}`;
        }
      }
    }

    if (!bankAccount) {
      newErrors.bankAccount = 'Please enter bank account number';
    } else if (bankAccount.length < 9 || bankAccount.length > 18) {
      newErrors.bankAccount = 'Invalid bank account number';
    }

    if (!ifscCode) {
      newErrors.ifscCode = 'Please enter IFSC code';
    } else if (!/^[A-Z]{4}0[A-Z0-9]{6}$/.test(ifscCode)) {
      newErrors.ifscCode = 'Invalid IFSC code format';
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
      await api.createWithdrawal({
        wallet_id: walletId,
        amount_paise: toPaise(parseFloat(amount)),
        description: `Withdrawal to ${bankAccount}`,
        reference: `${bankAccount}-${ifscCode}`,
      });

      setSuccess(true);
      setWalletId('');
      setAmount('');
      setBankAccount('');
      setIfscCode('');

      // Refetch wallets to update balance
      await fetchWallets();

      // Navigate back to dashboard after 2 seconds
      setTimeout(() => {
        navigate('/dashboard');
      }, 2000);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to process withdrawal');
      setIsLoading(false);
    }
  };

  const selectedWallet = wallets.find(w => w.id === walletId);

  return (
    <div className="min-h-screen bg-gray-50">
      {/* Navigation */}
      <nav className="bg-white shadow-sm">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="flex justify-between h-16 items-center">
            <h1 className="text-xl font-bold text-primary-600">Nivo Money</h1>
            <button onClick={() => navigate('/dashboard')} className="btn-secondary">
              Back to Dashboard
            </button>
          </div>
        </div>
      </nav>

      {/* Main Content */}
      <main className="max-w-2xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        <div className="card">
          <h2 className="text-2xl font-bold mb-6">Withdraw Money</h2>

          {error && (
            <div className="mb-4 p-4 bg-red-100 text-red-800 rounded-lg">
              {error}
            </div>
          )}

          {success && (
            <div className="mb-4 p-4 bg-green-100 text-green-800 rounded-lg">
              Withdrawal initiated successfully! Redirecting to dashboard...
            </div>
          )}

          <form onSubmit={handleSubmit} className="space-y-6">
            {/* Wallet Selection */}
            <div>
              <label htmlFor="wallet" className="block text-sm font-medium text-gray-700 mb-2">
                From Wallet
              </label>
              <select
                id="wallet"
                value={walletId}
                onChange={e => setWalletId(e.target.value)}
                className="input-field"
                disabled={isLoading}
              >
                <option value="">Select a wallet</option>
                {wallets
                  .filter(w => w.status === 'active')
                  .map(wallet => (
                    <option key={wallet.id} value={wallet.id}>
                      {wallet.type.toUpperCase()} - {formatCurrency(wallet.available_balance)}
                    </option>
                  ))}
              </select>
              {errors.walletId && (
                <p className="text-sm text-red-600 mt-1">{errors.walletId}</p>
              )}
              {selectedWallet && (
                <p className="text-sm text-gray-600 mt-1">
                  Available balance: {formatCurrency(selectedWallet.available_balance)}
                </p>
              )}
            </div>

            {/* Amount */}
            <div>
              <label htmlFor="amount" className="block text-sm font-medium text-gray-700 mb-2">
                Amount (₹)
              </label>
              <input
                type="number"
                id="amount"
                value={amount}
                onChange={e => setAmount(e.target.value)}
                className="input-field"
                placeholder="0.00"
                step="0.01"
                min="0"
                disabled={isLoading}
              />
              {errors.amount && (
                <p className="text-sm text-red-600 mt-1">{errors.amount}</p>
              )}
            </div>

            {/* Bank Account */}
            <div>
              <label htmlFor="bankAccount" className="block text-sm font-medium text-gray-700 mb-2">
                Bank Account Number
              </label>
              <input
                type="text"
                id="bankAccount"
                value={bankAccount}
                onChange={e => setBankAccount(e.target.value)}
                className="input-field"
                placeholder="Enter your bank account number"
                disabled={isLoading}
              />
              {errors.bankAccount && (
                <p className="text-sm text-red-600 mt-1">{errors.bankAccount}</p>
              )}
            </div>

            {/* IFSC Code */}
            <div>
              <label htmlFor="ifscCode" className="block text-sm font-medium text-gray-700 mb-2">
                IFSC Code
              </label>
              <input
                type="text"
                id="ifscCode"
                value={ifscCode}
                onChange={e => setIfscCode(e.target.value.toUpperCase())}
                className="input-field"
                placeholder="e.g., SBIN0001234"
                disabled={isLoading}
                maxLength={11}
              />
              {errors.ifscCode && (
                <p className="text-sm text-red-600 mt-1">{errors.ifscCode}</p>
              )}
            </div>

            {/* Submit Button */}
            <button
              type="submit"
              className="btn-primary w-full"
              disabled={isLoading}
            >
              {isLoading ? 'Processing...' : 'Withdraw Money'}
            </button>
          </form>
        </div>
      </main>
    </div>
  );
}

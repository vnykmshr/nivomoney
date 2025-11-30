import { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import { useWalletStore } from '../stores/walletStore';
import { api } from '../lib/api';
import { formatCurrency, toPaise } from '../lib/utils';

interface SavedBankAccount {
  id: string;
  accountNumber: string;
  ifscCode: string;
  bankName: string;
  accountHolderName: string;
}

export function Withdraw() {
  const navigate = useNavigate();
  const { wallets, fetchWallets } = useWalletStore();
  const [walletId, setWalletId] = useState('');
  const [amount, setAmount] = useState('');
  const [bankAccount, setBankAccount] = useState('');
  const [ifscCode, setIfscCode] = useState('');
  const [bankName, setBankName] = useState('');
  const [accountHolderName, setAccountHolderName] = useState('');
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [success, setSuccess] = useState(false);
  const [errors, setErrors] = useState<Record<string, string>>({});
  const [savedAccounts, setSavedAccounts] = useState<SavedBankAccount[]>([]);
  const [selectedSavedAccount, setSelectedSavedAccount] = useState<string>('');
  const [saveThisAccount, setSaveThisAccount] = useState(false);

  // Load saved bank accounts from localStorage
  useEffect(() => {
    const saved = localStorage.getItem('saved_bank_accounts');
    if (saved) {
      try {
        setSavedAccounts(JSON.parse(saved));
      } catch (err) {
        console.error('Failed to parse saved accounts:', err);
      }
    }
  }, []);

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

  // When a saved account is selected, populate the form
  useEffect(() => {
    if (selectedSavedAccount) {
      const account = savedAccounts.find(a => a.id === selectedSavedAccount);
      if (account) {
        setBankAccount(account.accountNumber);
        setIfscCode(account.ifscCode);
        setBankName(account.bankName);
        setAccountHolderName(account.accountHolderName);
      }
    }
  }, [selectedSavedAccount, savedAccounts]);

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

    if (saveThisAccount) {
      if (!bankName) {
        newErrors.bankName = 'Bank name is required to save this account';
      }
      if (!accountHolderName) {
        newErrors.accountHolderName = 'Account holder name is required to save this account';
      }
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
      // Save bank account if requested
      if (saveThisAccount && bankName && accountHolderName) {
        const newAccount: SavedBankAccount = {
          id: Date.now().toString(),
          accountNumber: bankAccount,
          ifscCode,
          bankName,
          accountHolderName,
        };
        const updatedAccounts = [...savedAccounts, newAccount];
        setSavedAccounts(updatedAccounts);
        localStorage.setItem('saved_bank_accounts', JSON.stringify(updatedAccounts));
      }

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
      setBankName('');
      setAccountHolderName('');
      setSelectedSavedAccount('');
      setSaveThisAccount(false);

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

            {/* Saved Bank Accounts */}
            {savedAccounts.length > 0 && (
              <div>
                <label htmlFor="savedAccount" className="block text-sm font-medium text-gray-700 mb-2">
                  Use Saved Bank Account
                </label>
                <select
                  id="savedAccount"
                  value={selectedSavedAccount}
                  onChange={e => setSelectedSavedAccount(e.target.value)}
                  className="input-field"
                  disabled={isLoading}
                >
                  <option value="">Enter new bank account details</option>
                  {savedAccounts.map(account => (
                    <option key={account.id} value={account.id}>
                      {account.bankName} - {account.accountNumber.slice(-4).padStart(account.accountNumber.length, '*')}
                    </option>
                  ))}
                </select>
              </div>
            )}

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

            {/* Bank Name (optional unless saving) */}
            <div>
              <label htmlFor="bankName" className="block text-sm font-medium text-gray-700 mb-2">
                Bank Name {saveThisAccount && <span className="text-red-600">*</span>}
              </label>
              <input
                type="text"
                id="bankName"
                value={bankName}
                onChange={e => setBankName(e.target.value)}
                className="input-field"
                placeholder="e.g., State Bank of India"
                disabled={isLoading}
              />
              {errors.bankName && (
                <p className="text-sm text-red-600 mt-1">{errors.bankName}</p>
              )}
            </div>

            {/* Account Holder Name (optional unless saving) */}
            <div>
              <label htmlFor="accountHolderName" className="block text-sm font-medium text-gray-700 mb-2">
                Account Holder Name {saveThisAccount && <span className="text-red-600">*</span>}
              </label>
              <input
                type="text"
                id="accountHolderName"
                value={accountHolderName}
                onChange={e => setAccountHolderName(e.target.value)}
                className="input-field"
                placeholder="As per bank records"
                disabled={isLoading}
              />
              {errors.accountHolderName && (
                <p className="text-sm text-red-600 mt-1">{errors.accountHolderName}</p>
              )}
            </div>

            {/* Save This Account Checkbox */}
            <div className="flex items-center">
              <input
                type="checkbox"
                id="saveAccount"
                checked={saveThisAccount}
                onChange={e => setSaveThisAccount(e.target.checked)}
                className="h-4 w-4 text-primary-600 focus:ring-primary-500 border-gray-300 rounded"
                disabled={isLoading}
              />
              <label htmlFor="saveAccount" className="ml-2 block text-sm text-gray-700">
                Save this bank account for future withdrawals
              </label>
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

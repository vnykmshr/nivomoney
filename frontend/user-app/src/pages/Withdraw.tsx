import { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import { useWalletStore } from '../stores/walletStore';
import { api } from '../lib/api';
import { formatCurrency, toPaise } from '../lib/utils';
import { AppLayout } from '../components';
import {
  Alert,
  Button,
  Card,
  FormField,
  Input,
  Checkbox,
} from '../../../shared/components';

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
    <AppLayout title="Withdraw" showBack>
      <div className="max-w-md mx-auto px-4 py-6 space-y-6">
        {error && (
          <Alert variant="error" onDismiss={() => setError(null)}>
            {error}
          </Alert>
        )}

        {success && (
          <Alert variant="success">
            Withdrawal initiated successfully! Redirecting to dashboard...
          </Alert>
        )}

        <Card padding="lg">

          <form onSubmit={handleSubmit} className="space-y-6">
            {/* Wallet Selection */}
            <FormField
              label="From Wallet"
              htmlFor="wallet"
              error={errors.walletId}
              hint={selectedWallet ? `Available balance: ${formatCurrency(selectedWallet.available_balance)}` : undefined}
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
                      {wallet.type.toUpperCase()} - {formatCurrency(wallet.available_balance)}
                    </option>
                  ))}
              </select>
            </FormField>

            {/* Amount */}
            <FormField label="Amount (₹)" htmlFor="amount" error={errors.amount} required>
              <Input
                type="number"
                id="amount"
                value={amount}
                onChange={e => setAmount(e.target.value)}
                placeholder="0.00"
                step="0.01"
                min="0"
                disabled={isLoading}
                error={!!errors.amount}
              />
            </FormField>

            {/* Saved Bank Accounts */}
            {savedAccounts.length > 0 && (
              <FormField label="Use Saved Bank Account" htmlFor="savedAccount">
                <select
                  id="savedAccount"
                  value={selectedSavedAccount}
                  onChange={e => setSelectedSavedAccount(e.target.value)}
                  className="w-full h-10 px-3 pr-10 text-sm appearance-none rounded-[var(--radius-input)] border border-[var(--input-border)] bg-[var(--input-bg)] text-[var(--input-text)] focus:border-[var(--input-border-focus)] focus:outline-none focus:[box-shadow:var(--focus-ring)] disabled:opacity-50 disabled:cursor-not-allowed"
                  disabled={isLoading}
                >
                  <option value="">Enter new bank account details</option>
                  {savedAccounts.map(account => (
                    <option key={account.id} value={account.id}>
                      {account.bankName} - {account.accountNumber.slice(-4).padStart(account.accountNumber.length, '*')}
                    </option>
                  ))}
                </select>
              </FormField>
            )}

            {/* Bank Account */}
            <FormField label="Bank Account Number" htmlFor="bankAccount" error={errors.bankAccount} required>
              <Input
                type="text"
                id="bankAccount"
                value={bankAccount}
                onChange={e => setBankAccount(e.target.value)}
                placeholder="Enter your bank account number"
                disabled={isLoading}
                error={!!errors.bankAccount}
              />
            </FormField>

            {/* IFSC Code */}
            <FormField label="IFSC Code" htmlFor="ifscCode" error={errors.ifscCode} required>
              <Input
                type="text"
                id="ifscCode"
                value={ifscCode}
                onChange={e => setIfscCode(e.target.value.toUpperCase())}
                placeholder="e.g., SBIN0001234"
                disabled={isLoading}
                maxLength={11}
                error={!!errors.ifscCode}
              />
            </FormField>

            {/* Bank Name (optional unless saving) */}
            <FormField
              label="Bank Name"
              htmlFor="bankName"
              error={errors.bankName}
              required={saveThisAccount}
            >
              <Input
                type="text"
                id="bankName"
                value={bankName}
                onChange={e => setBankName(e.target.value)}
                placeholder="e.g., State Bank of India"
                disabled={isLoading}
                error={!!errors.bankName}
              />
            </FormField>

            {/* Account Holder Name (optional unless saving) */}
            <FormField
              label="Account Holder Name"
              htmlFor="accountHolderName"
              error={errors.accountHolderName}
              required={saveThisAccount}
            >
              <Input
                type="text"
                id="accountHolderName"
                value={accountHolderName}
                onChange={e => setAccountHolderName(e.target.value)}
                placeholder="As per bank records"
                disabled={isLoading}
                error={!!errors.accountHolderName}
              />
            </FormField>

            {/* Save This Account Checkbox */}
            <Checkbox
              id="saveAccount"
              checked={saveThisAccount}
              onChange={e => setSaveThisAccount(e.target.checked)}
              disabled={isLoading}
              label="Save this bank account for future withdrawals"
            />

            {/* Submit Button */}
            <Button
              type="submit"
              className="w-full"
              loading={isLoading}
            >
              Withdraw Money
            </Button>
          </form>
        </Card>
      </div>
    </AppLayout>
  );
}

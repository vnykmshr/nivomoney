import type { Wallet } from '../types';
import { formatCurrency } from '../lib/utils';

interface WalletCardProps {
  wallet: Wallet;
  isSelected: boolean;
  onClick: () => void;
}

export function WalletCard({ wallet, isSelected, onClick }: WalletCardProps) {
  const getStatusColor = (status: string) => {
    const colors: Record<string, string> = {
      active: 'bg-green-100 text-green-800',
      inactive: 'bg-gray-100 text-gray-800',
      frozen: 'bg-yellow-100 text-yellow-800',
      closed: 'bg-red-100 text-red-800',
    };
    return colors[status] || 'bg-gray-100 text-gray-800';
  };

  return (
    <div
      onClick={onClick}
      className={`card cursor-pointer transition-all hover:shadow-md ${
        isSelected ? 'ring-2 ring-primary-500' : ''
      }`}
    >
      <div className="flex justify-between items-start mb-3">
        <div className="text-sm font-medium text-gray-700">
          {wallet.currency} Wallet
        </div>
        <span
          className={`inline-block px-2 py-1 text-xs font-semibold rounded ${getStatusColor(wallet.status)}`}
        >
          {wallet.status}
        </span>
      </div>

      <div className="mb-2">
        <div className="text-sm text-gray-600">Balance</div>
        <div className="text-2xl font-bold text-gray-900">{formatCurrency(wallet.balance)}</div>
      </div>

      <div>
        <div className="text-sm text-gray-600">Available Balance</div>
        <div className="text-lg font-semibold text-gray-700">
          {formatCurrency(wallet.available_balance)}
        </div>
      </div>

      <div className="mt-3 pt-3 border-t border-gray-200">
        <div className="text-xs text-gray-500">
          {wallet.currency} • ID: {wallet.id.slice(0, 8)}...
        </div>
      </div>
    </div>
  );
}

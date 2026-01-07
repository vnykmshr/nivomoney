/**
 * Toast Notification Component
 * Provides transient notifications that auto-dismiss
 */

import { createContext, useContext, useState, useCallback, useEffect } from 'react';
import type { ReactNode } from 'react';
import { cn } from '../../lib/utils';

export type ToastVariant = 'success' | 'warning' | 'error' | 'info';

export interface ToastData {
  id: string;
  variant: ToastVariant;
  title?: string;
  message: string;
  duration?: number;
}

export interface ToastProps {
  toast: ToastData;
  onDismiss: (id: string) => void;
}

const defaultIcons: Record<ToastVariant, ReactNode> = {
  success: (
    <svg className="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor" aria-hidden="true">
      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z" />
    </svg>
  ),
  warning: (
    <svg className="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor" aria-hidden="true">
      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z" />
    </svg>
  ),
  error: (
    <svg className="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor" aria-hidden="true">
      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M10 14l2-2m0 0l2-2m-2 2l-2-2m2 2l2 2m7-2a9 9 0 11-18 0 9 9 0 0118 0z" />
    </svg>
  ),
  info: (
    <svg className="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor" aria-hidden="true">
      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
    </svg>
  ),
};

/**
 * Individual Toast component
 */
export function Toast({ toast, onDismiss }: ToastProps) {
  const { id, variant, title, message, duration = 5000 } = toast;

  useEffect(() => {
    if (duration > 0) {
      const timer = setTimeout(() => {
        onDismiss(id);
      }, duration);
      return () => clearTimeout(timer);
    }
  }, [id, duration, onDismiss]);

  return (
    <div
      role="alert"
      aria-live="assertive"
      aria-atomic="true"
      className={cn(
        'relative flex items-start gap-3 p-4 min-w-[320px] max-w-[420px]',
        'rounded-[var(--radius-card)]',
        'border shadow-lg',
        'animate-[slideIn_0.3s_ease-out]',

        // Variant styles
        variant === 'success' && [
          'bg-[var(--surface-success)]',
          'border-[var(--color-success-200)]',
          'text-[var(--text-success)]',
        ],
        variant === 'warning' && [
          'bg-[var(--surface-warning)]',
          'border-[var(--color-warning-200)]',
          'text-[var(--text-warning)]',
        ],
        variant === 'error' && [
          'bg-[var(--surface-error)]',
          'border-[var(--color-error-200)]',
          'text-[var(--text-error)]',
        ],
        variant === 'info' && [
          'bg-[var(--surface-brand-subtle)]',
          'border-[var(--color-primary-200)]',
          'text-[var(--text-link)]',
        ]
      )}
    >
      <div className="flex-shrink-0">{defaultIcons[variant]}</div>
      <div className="flex-1 min-w-0">
        {title && (
          <h4 className="font-semibold mb-1">{title}</h4>
        )}
        <p className="text-sm opacity-90">{message}</p>
      </div>
      <button
        onClick={() => onDismiss(id)}
        className="flex-shrink-0 opacity-70 hover:opacity-100 transition-opacity p-1"
        aria-label="Dismiss notification"
      >
        <svg className="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" aria-hidden="true">
          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
        </svg>
      </button>
    </div>
  );
}

/**
 * Toast Context
 */
interface ToastContextType {
  toasts: ToastData[];
  addToast: (toast: Omit<ToastData, 'id'>) => void;
  removeToast: (id: string) => void;
  success: (message: string, title?: string) => void;
  warning: (message: string, title?: string) => void;
  error: (message: string, title?: string) => void;
  info: (message: string, title?: string) => void;
}

const ToastContext = createContext<ToastContextType | null>(null);

/**
 * Toast Provider - Wrap your app with this to enable toasts
 */
export interface ToastProviderProps {
  children: ReactNode;
  position?: 'top-right' | 'top-left' | 'bottom-right' | 'bottom-left' | 'top-center' | 'bottom-center';
  maxToasts?: number;
}

export function ToastProvider({
  children,
  position = 'top-right',
  maxToasts = 5,
}: ToastProviderProps) {
  const [toasts, setToasts] = useState<ToastData[]>([]);

  const removeToast = useCallback((id: string) => {
    setToasts(prev => prev.filter(t => t.id !== id));
  }, []);

  const addToast = useCallback((toast: Omit<ToastData, 'id'>) => {
    const id = `toast-${Date.now()}-${Math.random().toString(36).substring(2, 9)}`;
    setToasts(prev => {
      const newToasts = [...prev, { ...toast, id }];
      // Keep only the most recent maxToasts
      return newToasts.slice(-maxToasts);
    });
  }, [maxToasts]);

  // Convenience methods
  const success = useCallback((message: string, title?: string) => {
    addToast({ variant: 'success', message, title });
  }, [addToast]);

  const warning = useCallback((message: string, title?: string) => {
    addToast({ variant: 'warning', message, title });
  }, [addToast]);

  const error = useCallback((message: string, title?: string) => {
    addToast({ variant: 'error', message, title });
  }, [addToast]);

  const info = useCallback((message: string, title?: string) => {
    addToast({ variant: 'info', message, title });
  }, [addToast]);

  const positionClasses = {
    'top-right': 'top-4 right-4',
    'top-left': 'top-4 left-4',
    'bottom-right': 'bottom-4 right-4',
    'bottom-left': 'bottom-4 left-4',
    'top-center': 'top-4 left-1/2 -translate-x-1/2',
    'bottom-center': 'bottom-4 left-1/2 -translate-x-1/2',
  };

  return (
    <ToastContext.Provider value={{ toasts, addToast, removeToast, success, warning, error, info }}>
      {children}
      {/* Toast container */}
      <div
        className={cn(
          'fixed z-[100] flex flex-col gap-2',
          positionClasses[position]
        )}
        aria-label="Notifications"
      >
        {toasts.map(toast => (
          <Toast key={toast.id} toast={toast} onDismiss={removeToast} />
        ))}
      </div>
    </ToastContext.Provider>
  );
}

/**
 * Hook to use toast notifications
 */
export function useToast() {
  const context = useContext(ToastContext);
  if (!context) {
    throw new Error('useToast must be used within a ToastProvider');
  }
  return context;
}

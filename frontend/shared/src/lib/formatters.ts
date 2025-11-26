/**
 * Shared Formatting Utilities
 * Common formatters for currency, dates, and other data
 */

/**
 * Format amount in paise to INR currency
 * @param amountPaise - Amount in paise (1 INR = 100 paise)
 * @returns Formatted currency string (e.g., "₹1,234.56")
 */
export function formatCurrency(amountPaise: number): string {
  const amount = amountPaise / 100;
  return new Intl.NumberFormat('en-IN', {
    style: 'currency',
    currency: 'INR',
    minimumFractionDigits: 2,
  }).format(amount);
}

/**
 * Convert INR amount to paise
 * @param amount - Amount in INR
 * @returns Amount in paise
 */
export function toPaise(amount: number): number {
  return Math.round(amount * 100);
}

/**
 * Format paise amount to rupees without currency symbol
 * @param amountPaise - Amount in paise
 * @returns Formatted amount (e.g., "1,234.56")
 */
export function formatAmount(amountPaise: number): string {
  const amount = amountPaise / 100;
  return new Intl.NumberFormat('en-IN', {
    minimumFractionDigits: 2,
    maximumFractionDigits: 2,
  }).format(amount);
}

/**
 * Format date to human-readable format
 * @param dateString - ISO date string
 * @returns Formatted date (e.g., "Dec 15, 2023, 10:30 AM")
 */
export function formatDate(dateString: string): string {
  const date = new Date(dateString);
  return new Intl.DateTimeFormat('en-IN', {
    year: 'numeric',
    month: 'short',
    day: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
  }).format(date);
}

/**
 * Format date to short format without time
 * @param dateString - ISO date string
 * @returns Formatted date (e.g., "Dec 15, 2023")
 */
export function formatDateShort(dateString: string): string {
  const date = new Date(dateString);
  return new Intl.DateTimeFormat('en-IN', {
    year: 'numeric',
    month: 'short',
    day: 'numeric',
  }).format(date);
}

/**
 * Format date to relative time (e.g., "2 hours ago")
 * @param dateString - ISO date string
 * @returns Relative time string
 */
export function formatRelativeTime(dateString: string): string {
  const date = new Date(dateString);
  const now = new Date();
  const diffInSeconds = Math.floor((now.getTime() - date.getTime()) / 1000);

  if (diffInSeconds < 60) return 'Just now';
  if (diffInSeconds < 3600) return `${Math.floor(diffInSeconds / 60)} min ago`;
  if (diffInSeconds < 86400) return `${Math.floor(diffInSeconds / 3600)} hours ago`;
  if (diffInSeconds < 604800) return `${Math.floor(diffInSeconds / 86400)} days ago`;

  return formatDate(dateString);
}

/**
 * Format phone number for display
 * @param phone - Phone number (with or without +91 prefix)
 * @returns Formatted phone (e.g., "+91 98765 43210")
 */
export function formatPhone(phone: string): string {
  // Remove all non-digit characters
  const digits = phone.replace(/\D/g, '');

  // If starts with 91 and is 12 digits, format as Indian mobile
  if (digits.startsWith('91') && digits.length === 12) {
    const number = digits.substring(2);
    return `+91 ${number.substring(0, 5)} ${number.substring(5)}`;
  }

  // If 10 digits, assume Indian mobile without country code
  if (digits.length === 10) {
    return `+91 ${digits.substring(0, 5)} ${digits.substring(5)}`;
  }

  // Return as-is if doesn't match expected format
  return phone;
}

/**
 * Format PAN for display (mask middle digits)
 * @param pan - PAN number
 * @returns Masked PAN (e.g., "ABCD******")
 */
export function formatPANMasked(pan: string): string {
  if (pan.length < 6) return pan;
  return pan.substring(0, 4) + '******';
}

/**
 * Format Aadhaar for display (mask middle digits)
 * @param aadhaar - Aadhaar number
 * @returns Masked Aadhaar (e.g., "XXXX XXXX 1234")
 */
export function formatAadhaarMasked(aadhaar: string): string {
  const digits = aadhaar.replace(/\D/g, '');
  if (digits.length !== 12) return aadhaar;
  return `XXXX XXXX ${digits.substring(8)}`;
}

/**
 * Capitalize first letter of each word
 * @param text - Input text
 * @returns Capitalized text
 */
export function capitalizeWords(text: string): string {
  return text
    .split(' ')
    .map((word) => word.charAt(0).toUpperCase() + word.slice(1).toLowerCase())
    .join(' ');
}

/**
 * Truncate text to specified length
 * @param text - Input text
 * @param maxLength - Maximum length
 * @returns Truncated text with ellipsis
 */
export function truncate(text: string, maxLength: number): string {
  if (text.length <= maxLength) return text;
  return text.substring(0, maxLength - 3) + '...';
}

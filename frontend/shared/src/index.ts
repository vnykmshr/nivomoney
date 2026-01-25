/**
 * @nivo/shared
 * Shared utilities, types, and API client for Nivo Money applications
 */

// Export all types
export * from './types';

// Export API client
export { BaseApiClient } from './lib/apiClient';
export type { ApiClientConfig } from './lib/apiClient';

// Export validation utilities
export * from './lib/validation';

// Export formatters
export * from './lib/formatters';

// Export constants
export * from './lib/constants';

// Export phone utilities
export * from './utils/phone';

// Export security utilities
export * from './lib/security';

// Export UI components
export * from '../components';

# @nivo/shared

Shared utilities, types, and API client for Nivo Money applications.

## Overview

This package provides common functionality used across both the user-app and admin-app:

- **Type Definitions**: User, Wallet, Transaction, KYC, and API types
- **API Client**: Base HTTP client with authentication and error handling
- **Validation**: Email, phone, PAN, Aadhaar, password validators
- **Formatters**: Currency, date, phone number formatters
- **Constants**: Shared constants for statuses, types, and configuration

## Installation

This package is part of the Nivo Money monorepo and is installed automatically via npm workspaces.

```bash
# From the project root
npm install
```

## Usage

### Importing Types

```typescript
import type { User, Wallet, Transaction, KYCInfo } from '@nivo/shared';
```

### Using the Base API Client

```typescript
import { BaseApiClient } from '@nivo/shared';

class MyApiClient extends BaseApiClient {
  constructor() {
    super({
      baseURL: 'http://localhost:8000',
      timeout: 15000,
      getToken: () => localStorage.getItem('auth_token'),
      onUnauthorized: () => {
        localStorage.removeItem('auth_token');
        window.location.href = '/login';
      },
    });
  }

  // Add your API methods
  async getUsers() {
    return this.get<User[]>('/api/v1/users');
  }
}
```

### Using Validation Functions

```typescript
import { isValidEmail, isValidPhone, isValidPAN } from '@nivo/shared';

if (!isValidEmail(email)) {
  console.error('Invalid email format');
}

if (!isValidPhone(phone)) {
  console.error('Invalid Indian phone number');
}

if (!isValidPAN(pan)) {
  console.error('Invalid PAN format');
}
```

### Using Formatters

```typescript
import { formatCurrency, formatDate, formatPhone } from '@nivo/shared';

const amount = formatCurrency(123456); // "₹1,234.56"
const date = formatDate('2023-12-15T10:30:00Z'); // "Dec 15, 2023, 10:30 AM"
const phone = formatPhone('9876543210'); // "+91 98765 43210"
```

### Using Constants

```typescript
import {
  WALLET_TYPES,
  STATUS_COLORS,
  getStatusColor,
  DEFAULT_CURRENCY,
} from '@nivo/shared';

const walletType = WALLET_TYPES.SAVINGS;
const statusColor = getStatusColor('completed'); // "bg-green-100 text-green-800"
```

## Package Structure

```
shared/
├── src/
│   ├── types/
│   │   └── index.ts          # Type definitions
│   ├── lib/
│   │   ├── apiClient.ts      # Base API client
│   │   ├── validation.ts     # Validation utilities
│   │   ├── formatters.ts     # Formatting utilities
│   │   └── constants.ts      # Shared constants
│   ├── utils/
│   │   └── phone.ts          # Phone number utilities
│   └── index.ts              # Main export file
├── package.json
├── tsconfig.json
└── README.md
```

## Development

### Building the Package

```bash
# Build for production
npm run build

# Watch mode for development
npm run dev

# Type checking
npm run typecheck
```

### Adding New Exports

1. Add your code to the appropriate directory (`types`, `lib`, or `utils`)
2. Export it from `src/index.ts`
3. Rebuild the package with `npm run build`
4. The new exports will be available to both user-app and admin-app

## Best Practices

### What Should Go in Shared?

✅ **DO Include**:
- Common type definitions (User, Wallet, Transaction)
- Shared validation logic (email, phone, PAN validation)
- Formatting utilities (currency, dates)
- Constants (API URLs, status values)
- Base API client functionality

❌ **DON'T Include**:
- React components (different UX per app)
- App-specific business logic
- Route definitions
- State management code
- App-specific styling

### Version Management

- Increment version when making breaking changes
- Document changes in commit messages
- Both apps should use the same version

### Type Safety

- Always export types, not just interfaces
- Use TypeScript strict mode
- Avoid using `any` - use `unknown` instead

## Dependencies

### Runtime Dependencies
- `axios` - HTTP client for API requests

### Dev Dependencies
- `typescript` - TypeScript compiler
- `tsup` - Build tool for TypeScript packages
- `@types/node` - Node.js type definitions

## License

UNLICENSED - Private package for Nivo Money internal use only.

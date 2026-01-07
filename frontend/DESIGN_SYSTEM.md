# Nivo Money Design System

A CSS variables-based design system providing consistent theming across all Nivo Money frontend applications.

## Architecture

```
shared/styles/tokens/
├── index.css        # Entry point - import order matters
├── primitives.css   # Raw values (colors, spacing, typography)
├── semantic.css     # Meaningful tokens mapped to primitives
└── derived.css      # States, dark mode overrides
```

### Token Hierarchy

```
Primitives → Semantic → Derived
     ↑            ↑          ↑
   Raw values   Meaning   States/Modes
```

**Rule**: Components should only use semantic or derived tokens, never primitives directly.

## Quick Start

Import in your app's `index.css`:

```css
@import "tailwindcss";
@import "../../shared/styles/tokens/index.css";
```

Use tokens in components:

```tsx
// React component using design tokens
<button className="bg-[var(--button-primary-bg)] text-[var(--button-primary-text)]">
  Click me
</button>
```

## Color Tokens

### Text Colors
| Token | Light Mode | Usage |
|-------|------------|-------|
| `--text-primary` | neutral-900 | Main content |
| `--text-secondary` | neutral-600 | Secondary content |
| `--text-muted` | neutral-400 | Hints, placeholders |
| `--text-inverse` | white | Text on dark backgrounds |
| `--text-link` | primary-600 | Link text |
| `--text-success` | success-700 | Success messages |
| `--text-warning` | warning-700 | Warning messages |
| `--text-error` | error-600 | Error messages |

### Surface Colors
| Token | Light Mode | Usage |
|-------|------------|-------|
| `--surface-page` | neutral-50 | Page background |
| `--surface-card` | white | Card backgrounds |
| `--surface-elevated` | white | Modals, dropdowns |
| `--surface-overlay` | rgba(0,0,0,0.5) | Modal overlays |
| `--surface-brand` | primary-500 | Brand elements |
| `--surface-brand-subtle` | primary-50 | Light brand backgrounds |
| `--surface-success` | success-50 | Success alerts |
| `--surface-warning` | warning-50 | Warning alerts |
| `--surface-error` | error-50 | Error alerts |

### Interactive Colors
| Token | Light Mode | Usage |
|-------|------------|-------|
| `--interactive-primary` | primary-600 | Primary buttons |
| `--interactive-primary-hover` | primary-700 | Primary hover |
| `--interactive-secondary` | neutral-100 | Secondary buttons |
| `--interactive-secondary-hover` | neutral-200 | Secondary hover |
| `--interactive-danger` | error-600 | Danger buttons |
| `--interactive-danger-hover` | error-700 | Danger hover |

### Border Colors
| Token | Light Mode | Usage |
|-------|------------|-------|
| `--border-default` | neutral-200 | Standard borders |
| `--border-subtle` | neutral-100 | Light borders |
| `--border-strong` | neutral-300 | Emphasized borders |
| `--border-focus` | primary-500 | Focus states |
| `--border-error` | error-500 | Error states |

## Component Tokens

Pre-configured tokens for common components:

### Buttons
```css
--button-primary-bg: var(--interactive-primary);
--button-primary-bg-hover: var(--interactive-primary-hover);
--button-primary-text: var(--text-inverse);
--button-secondary-bg: var(--interactive-secondary);
--button-secondary-bg-hover: var(--interactive-secondary-hover);
--button-secondary-text: var(--text-primary);
```

### Inputs
```css
--input-bg: var(--surface-card);
--input-border: var(--border-default);
--input-border-focus: var(--border-focus);
--input-text: var(--text-primary);
--input-placeholder: var(--text-muted);
```

### Cards
```css
--card-bg: var(--surface-card);
--card-border: var(--border-subtle);
--card-shadow: var(--shadow-md);
```

## Shape Tokens

### Border Radius
| Token | Value | Usage |
|-------|-------|-------|
| `--radius-button` | 8px | Buttons |
| `--radius-card` | 12px | Cards |
| `--radius-input` | 8px | Form inputs |
| `--radius-badge` | 9999px | Pills/badges |
| `--radius-modal` | 16px | Modal dialogs |

### Shadows
| Token | Usage |
|-------|-------|
| `--shadow-button` | Button elevation |
| `--shadow-card` | Card elevation |
| `--shadow-modal` | Modal elevation |
| `--shadow-dropdown` | Dropdown elevation |

## Focus States

```css
/* Default focus ring */
--focus-ring: 0 0 0 2px var(--surface-card), 0 0 0 4px var(--color-primary-500);

/* Error focus ring */
--focus-ring-error: 0 0 0 2px var(--surface-card), 0 0 0 4px var(--color-error-500);
```

Usage:
```css
focus:outline-none focus:[box-shadow:var(--focus-ring)]
```

## Dark Mode

Add `.dark` class to `<html>` element:

```html
<html class="dark">
```

Tokens automatically switch to dark mode values:
- Surface colors become dark neutrals
- Text colors become light neutrals
- Interactive colors become brighter for contrast
- Shadows become more subtle

## App-Specific Theming

### User App (Teal Theme)
Uses default tokens - Trust Teal (`#00ACB0`) as primary color.

### Admin App (Orange Theme)
Overrides primary color palette in `admin-app/src/index.css`:

```css
:root {
  --color-primary-50: #fff7ed;
  --color-primary-500: #f97316;
  --color-primary-600: #ea580c;
  /* ... */
}
```

Semantic tokens like `--interactive-primary` automatically use the overridden primaries.

## Shared Components

Located in `shared/components/ui/`:

| Component | Description |
|-----------|-------------|
| `Button` | Primary, secondary, outline, danger, ghost variants |
| `Input` | Text input with error states, icons, password toggle |
| `Modal` | Dialog with overlay, sizes, header/footer slots |
| `Alert` | Info, success, warning, error variants |
| `Card` | Container with header, body, footer sections |
| `Badge` | Status indicators with semantic colors |
| `Avatar` | User avatars with size variants |
| `Skeleton` | Loading placeholders |
| `Spinner` | Loading indicator |
| `Toast` | Notification system with provider/hook |
| `Table` | Data table with sorting |
| `Pagination` | Page navigation controls |

### Usage Example

```tsx
import { Button, Input, Card, Alert } from '@/shared/components/ui'

function MyForm() {
  return (
    <Card>
      <Card.Header>
        <Card.Title>Login</Card.Title>
      </Card.Header>
      <Card.Body>
        <Input placeholder="Email" type="email" />
        <Input placeholder="Password" type="password" />
      </Card.Body>
      <Card.Footer>
        <Button>Sign In</Button>
        <Button variant="secondary">Cancel</Button>
      </Card.Footer>
    </Card>
  )
}
```

## Testing

Tests are in `shared/components/__tests__/`. Run from user-app:

```bash
cd user-app
npm test          # Watch mode
npm run test:run  # Single run
npm run test:coverage  # With coverage
```

## Best Practices

1. **Use semantic tokens** - Never use primitives like `--color-neutral-200` directly
2. **Avoid hardcoded colors** - Use `bg-[var(--surface-page)]` not `bg-gray-50`
3. **Use component tokens** - Prefer `--button-primary-bg` over `--interactive-primary`
4. **Support dark mode** - Semantic tokens handle this automatically
5. **Use shared components** - Import from `@/shared/components/ui`

## Migration Guide

### From Hardcoded Tailwind Colors

```diff
- <div className="bg-gray-50 text-gray-500">
+ <div className="bg-[var(--surface-page)] text-[var(--text-muted)]">

- <button className="bg-teal-600 text-white">
+ <button className="bg-[var(--button-primary-bg)] text-[var(--button-primary-text)]">

- <div className="bg-black/50">
+ <div className="bg-[var(--surface-overlay)]">
```

### Common Mappings

| Hardcoded | Token |
|-----------|-------|
| `bg-gray-50` | `bg-[var(--surface-page)]` |
| `bg-white` | `bg-[var(--surface-card)]` |
| `text-gray-900` | `text-[var(--text-primary)]` |
| `text-gray-500` | `text-[var(--text-secondary)]` |
| `text-gray-400` | `text-[var(--text-muted)]` |
| `text-white` | `text-[var(--text-inverse)]` |
| `border-gray-200` | `border-[var(--border-default)]` |
| `bg-black/50` | `bg-[var(--surface-overlay)]` |
| `bg-teal-600` | `bg-[var(--interactive-primary)]` |
| `bg-red-600` | `bg-[var(--interactive-danger)]` |

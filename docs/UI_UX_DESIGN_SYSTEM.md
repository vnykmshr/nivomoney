# UI/UX Design System - Nivo Money

> **Version:** 1.0.0
> **Last Updated:** 2025-11-28
> **Status:** Active

## Overview

This document defines the comprehensive design system for Nivo Money's frontend applications, ensuring consistency, accessibility, and excellent user experience across both user-facing and administrative interfaces.

---

## 1. Color Schemes

### User App - Blue Theme (Trust & Security)

**Primary Blue**
- Purpose: Main brand color, primary actions, navigation
- Use cases: Buttons, links, headers, active states

```css
primary-50:  #eff6ff  /* Lightest backgrounds */
primary-100: #dbeafe  /* Light backgrounds, badges */
primary-200: #bfdbfe  /* Hover states for light elements */
primary-300: #93c5fd  /* Secondary UI elements */
primary-400: #60a5fa  /* Secondary interactions */
primary-500: #3b82f6  /* Main brand blue */
primary-600: #2563eb  /* Primary button default */
primary-700: #1d4ed8  /* Primary button hover */
primary-800: #1e40af  /* Primary button active */
primary-900: #1e3a8a  /* Dark text on light backgrounds */
primary-950: #172554  /* Darkest shades */
```

**Secondary Indigo**
- Purpose: Complementary accents, secondary actions
- Use cases: Secondary buttons, tags, highlights

```css
secondary-500: #6366f1  /* Main secondary color */
secondary-600: #4f46e5  /* Secondary actions */
```

### Admin App - Orange Theme (Energy & Action)

**Primary Orange**
- Purpose: Main admin brand, administrative actions
- Use cases: Buttons, admin navigation, action items

```css
primary-50:  #fff7ed  /* Lightest backgrounds */
primary-100: #ffedd5  /* Light backgrounds */
primary-500: #f97316  /* Main brand orange */
primary-600: #ea580c  /* Primary button default */
primary-700: #c2410c  /* Primary button hover */
primary-800: #9a3412  /* Primary button active */
```

**Accent Purple**
- Purpose: Admin badges, special indicators, priority items
- Use cases: Admin role badges, priority notifications

```css
accent-100: #f3e8ff  /* Admin badge background */
accent-500: #a855f7  /* Admin accent color */
accent-800: #6b21a8  /* Admin badge text */
```

### Semantic Colors (Both Apps)

**Success Green**
```css
success-50:  #f0fdf4  /* Success backgrounds */
success-100: #dcfce7  /* Success badges */
success-500: #22c55e  /* Success icons */
success-600: #16a34a  /* Success buttons */
success-800: #166534  /* Success text */
```

**Warning/Alert**
```css
warning-50:  #fffbeb  /* Warning backgrounds */
warning-100: #fef3c7  /* Warning badges */
warning-500: #f59e0b  /* Warning icons */
warning-800: #92400e  /* Warning text */
```

**Error/Danger**
```css
error-50:  #fef2f2  /* Error backgrounds */
error-100: #fee2e2  /* Error badges */
error-500: #ef4444  /* Error icons */
error-600: #dc2626  /* Danger buttons */
error-800: #991b1b  /* Error text */
```

---

## 2. Typography

### Font Stack
```css
font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', 'Roboto',
             'Oxygen', 'Ubuntu', 'Cantarell', 'Fira Sans', 'Droid Sans',
             'Helvetica Neue', sans-serif;
```

### Type Scale

```css
/* Headings */
text-5xl: 3rem (48px)     /* Hero headings */
text-4xl: 2.25rem (36px)  /* Page titles */
text-3xl: 1.875rem (30px) /* Section titles */
text-2xl: 1.5rem (24px)   /* Card headings */
text-xl:  1.25rem (20px)  /* Subheadings */
text-lg:  1.125rem (18px) /* Large body */

/* Body */
text-base: 1rem (16px)    /* Default body */
text-sm:   0.875rem (14px) /* Small text */
text-xs:   0.75rem (12px)  /* Micro text, badges */
```

### Font Weights
```css
font-bold:      700  /* Headings, emphasis */
font-semibold:  600  /* Buttons, labels */
font-medium:    500  /* Subheadings */
font-normal:    400  /* Body text */
```

### Line Heights
```css
leading-tight:  1.25  /* Headings */
leading-normal: 1.5   /* Body text */
leading-relaxed: 1.75 /* Paragraphs */
```

---

## 3. Spacing System

### Scale (Tailwind CSS)
```css
0:   0px
0.5: 0.125rem (2px)
1:   0.25rem  (4px)
1.5: 0.375rem (6px)
2:   0.5rem   (8px)
3:   0.75rem  (12px)
4:   1rem     (16px)
5:   1.25rem  (20px)
6:   1.5rem   (24px)
8:   2rem     (32px)
12:  3rem     (48px)
16:  4rem     (64px)
```

### Usage Guidelines
- **Component padding:** 1.5rem (p-6)
- **Button padding:** 0.5rem 1rem (py-2.5 px-4)
- **Input padding:** 0.625rem 0.75rem (py-2.5 px-3)
- **Section spacing:** 2rem - 3rem (py-8 to py-12)
- **Grid gaps:** 1rem - 1.5rem (gap-4 to gap-6)

---

## 4. Component Library

### Buttons

**Primary Button**
```css
.btn-primary {
  background: primary-600
  color: white
  padding: 0.625rem 1rem
  border-radius: 0.5rem
  font-weight: 600
  transition: all 200ms

  hover: primary-700, shadow-md
  active: primary-800, scale-95
  focus: ring-2 ring-primary-500 ring-offset-2
  disabled: opacity-50, cursor-not-allowed
}
```

**Secondary Button**
```css
.btn-secondary {
  background: gray-100
  color: gray-700
  padding: 0.625rem 1rem
  border-radius: 0.5rem
  font-weight: 600

  hover: gray-200, shadow-sm
}
```

**Outline Button**
```css
.btn-outline {
  border: 2px solid primary-600
  color: primary-600
  background: transparent

  hover: primary-50
}
```

**Danger Button**
```css
.btn-danger {
  background: error-600
  color: white

  hover: error-700
}
```

**Success Button (Admin only)**
```css
.btn-success {
  background: success-600
  color: white

  hover: success-700
}
```

### Form Inputs

**Input Field**
```css
.input-field {
  width: 100%
  padding: 0.625rem 0.75rem
  border: 1px solid gray-300
  border-radius: 0.5rem

  focus: ring-2 ring-primary-500, border-transparent
  placeholder: gray-400
  disabled: bg-gray-50, text-gray-500
}
```

**Error Input**
```css
.input-field-error {
  border-color: error-500
  focus: ring-error-500
}
```

**Label**
```css
.label {
  display: block
  font-size: 0.875rem
  font-weight: 500
  color: gray-700
  margin-bottom: 0.375rem
}
```

### Cards

**Standard Card**
```css
.card {
  background: white
  border-radius: 0.75rem
  box-shadow: 0 1px 3px rgba(0,0,0,0.05)
  padding: 1.5rem
  border: 1px solid gray-100
}
```

**Hover Card**
```css
.card-hover {
  /* extends .card */
  hover: shadow-md
  transition: shadow 200ms
}
```

### Badges

```css
.badge {
  display: inline-flex
  align-items: center
  padding: 0.125rem 0.625rem
  border-radius: 9999px
  font-size: 0.75rem
  font-weight: 500
}

.badge-success  { bg: success-100, text: success-800 }
.badge-warning  { bg: warning-100, text: warning-800 }
.badge-error    { bg: error-100, text: error-800 }
.badge-info     { bg: primary-100, text: primary-800 }
.badge-admin    { bg: accent-100, text: accent-800 }
```

### Alerts

```css
.alert {
  padding: 1rem
  border-radius: 0.5rem
  border-left: 4px solid
}

.alert-success  { bg: success-50, border: success-400, text: success-800 }
.alert-warning  { bg: warning-50, border: warning-400, text: warning-800 }
.alert-error    { bg: error-50, border: error-400, text: error-800 }
.alert-info     { bg: primary-50, border: primary-400, text: primary-800 }
```

---

## 5. Layout Guidelines

### Responsive Breakpoints
```css
sm:  640px   /* Small tablets */
md:  768px   /* Tablets */
lg:  1024px  /* Laptops */
xl:  1280px  /* Desktops */
2xl: 1536px  /* Large screens */
```

### Container Widths
```css
max-width: 1280px (max-w-7xl)  /* Main content container */
padding: 1rem (px-4) mobile
padding: 1.5rem (px-6) tablet
padding: 2rem (px-8) desktop
```

### Grid Systems
```css
/* Dashboard cards */
grid-cols-1 md:grid-cols-2 lg:grid-cols-3
gap-4 md:gap-6

/* Stats cards */
grid-cols-1 md:grid-cols-5
gap-4
```

---

## 6. Interaction States

### Transitions
```css
/* Standard transition */
transition-all duration-200 ease-in-out

/* Color transition only */
transition-colors duration-200

/* Shadow transition */
transition-shadow duration-200
```

### Hover Effects
- **Buttons:** Color darkening + shadow elevation
- **Cards:** Shadow elevation
- **Links:** Color change + underline
- **Icons:** Scale 110% + color change

### Active States
- **Buttons:** Scale 95% + darker color
- **Inputs:** Ring + border color change
- **Navigation:** Bold + color indicator

### Disabled States
- **Opacity:** 50%
- **Cursor:** not-allowed
- **Hover effects:** Disabled

---

## 7. Accessibility Guidelines

### Color Contrast
- **Body text on white:** Minimum 4.5:1 (WCAG AA)
- **Large text on white:** Minimum 3:1
- **Button text:** Always use sufficient contrast
  - White text on primary-600: ✅ Pass
  - Dark text on primary-50: ✅ Pass

### Focus Indicators
- **All interactive elements:** Visible focus ring
- **Ring style:** 2px solid, offset 2px
- **Ring color:** Matches element theme

### Keyboard Navigation
- **Tab order:** Logical flow
- **Skip links:** Implemented on complex pages
- **Enter/Space:** Activates buttons and links

### ARIA Labels
- **Icons without text:** aria-label required
- **Loading states:** aria-busy="true"
- **Error states:** aria-invalid="true"

---

## 8. Animation Principles

### Durations
```css
/* Micro-interactions */
duration-150: 150ms  (buttons, hover states)

/* Standard transitions */
duration-200: 200ms  (cards, modals)

/* Page transitions */
duration-300: 300ms  (route changes)
```

### Easing
```css
ease-in-out:  cubic-bezier(0.4, 0, 0.2, 1)  /* Standard */
ease-out:     cubic-bezier(0, 0, 0.2, 1)     /* Entrances */
ease-in:      cubic-bezier(0.4, 0, 1, 1)     /* Exits */
```

### Scale Effects
```css
hover: scale-105  /* Subtle emphasis */
active: scale-95  /* Button press feedback */
```

---

## 9. Shadow System

```css
/* Elevation levels */
shadow-sm:   0 1px 2px rgba(0, 0, 0, 0.05)   /* Cards at rest */
shadow:      0 1px 3px rgba(0, 0, 0, 0.1)    /* Default elevation */
shadow-md:   0 4px 6px rgba(0, 0, 0, 0.1)    /* Hover states */
shadow-lg:   0 10px 15px rgba(0, 0, 0, 0.1)  /* Modals */
shadow-xl:   0 20px 25px rgba(0, 0, 0, 0.1)  /* Overlays */
```

---

## 10. Icon System

### Icon Library
- **Source:** Heroicons (inline SVG)
- **Sizes:**
  - Small: 1rem (16px) - w-4 h-4
  - Medium: 1.25rem (20px) - w-5 h-5
  - Large: 1.5rem (24px) - w-6 h-6
  - XL: 2rem (32px) - w-8 h-8

### Usage
```tsx
// Success icon
<svg className="w-5 h-5 text-success-500" fill="currentColor" viewBox="0 0 20 20">
  <path fillRule="evenodd" d="..." clipRule="evenodd" />
</svg>
```

---

## 11. Page Templates

### User App Pages
- **Landing Page:** Hero + Features + CTA
- **Auth Pages:** Centered card on gradient background
- **Dashboard:** Nav + Stats grid + Content sections
- **Form Pages:** Centered form with sidebar info

### Admin App Pages
- **Admin Login:** Centered card with security notices
- **Admin Dashboard:** Stats grid + Tabs + Action cards
- **Detail Pages:** Header + Info grid + Action buttons

---

## 12. Implementation Checklist

### For New Components
- [ ] Uses design system colors (primary, success, error, etc.)
- [ ] Follows spacing scale (p-6, gap-4, etc.)
- [ ] Has all interaction states (hover, active, focus, disabled)
- [ ] Includes focus indicators for accessibility
- [ ] Uses consistent typography scale
- [ ] Has smooth transitions (duration-200)
- [ ] Responsive across all breakpoints
- [ ] Color contrast meets WCAG AA standards

### For New Pages
- [ ] Consistent navigation header
- [ ] Proper content container (max-w-7xl)
- [ ] Responsive grid layouts
- [ ] Loading states implemented
- [ ] Error states handled
- [ ] Breadcrumbs/back navigation where needed
- [ ] Page title follows hierarchy

---

## 13. Design Tokens Reference

### Quick Reference Table

| Element | User App | Admin App |
|---------|----------|-----------|
| Primary Color | Blue (#2563eb) | Orange (#ea580c) |
| Button Hover | Blue-700 | Orange-700 |
| Focus Ring | Blue-500 | Orange-500 |
| Success | Green (#16a34a) | Green (#16a34a) |
| Error | Red (#dc2626) | Red (#dc2626) |
| Warning | Amber (#d97706) | Yellow (#ca8a04) |
| Badge Accent | Indigo | Purple (#9333ea) |

---

## 14. Best Practices

### Do's ✅
- Use semantic colors (success, warning, error) for status
- Maintain consistent spacing throughout pages
- Provide clear visual feedback for interactions
- Use loading states for async operations
- Keep button labels concise and action-oriented
- Use proper heading hierarchy (h1 > h2 > h3)

### Don'ts ❌
- Don't mix color systems between apps
- Don't skip hover/focus states on interactive elements
- Don't use color as the only indicator of state
- Don't nest cards more than 2 levels deep
- Don't use custom shadows outside the system
- Don't override Tailwind classes with inline styles

---

## 15. Version History

| Version | Date | Changes |
|---------|------|---------|
| 1.0.0 | 2025-11-28 | Initial design system with blue/orange theme separation |

---

**Maintained by:** Engineering Team
**Contact:** For questions or suggestions, create an issue on GitHub

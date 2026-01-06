# Nivo Documentation

This directory contains the documentation for the Nivo neobank platform, built with Jekyll and hosted on GitHub Pages.

## Live Documentation

Visit: [docs.nivomoney.com](https://docs.nivomoney.com)

## Local Development

### Prerequisites

- Ruby 2.7+ with bundler
- OR Docker

### Option 1: Using Ruby

```bash
# Navigate to docs directory
cd docs

# Install dependencies
bundle install

# Start local server
bundle exec jekyll serve

# Open http://localhost:4000
```

### Option 2: Using Docker

```bash
cd docs

docker run --rm \
  --volume="$PWD:/srv/jekyll:Z" \
  -p 4000:4000 \
  jekyll/jekyll:4.3 \
  jekyll serve
```

## Structure

```
docs/
├── _config.yml           # Jekyll configuration
├── _sass/
│   └── color_schemes/
│       └── nivo.scss    # Custom Nivo theme colors
├── assets/
│   └── images/
│       └── logo.svg     # Documentation logo
├── index.md             # Home page
├── QUICKSTART.md        # Quick start guide
├── DEVELOPMENT.md       # Development guide
├── END_TO_END_FLOWS.md  # API flow documentation
├── UI_UX_DESIGN_SYSTEM.md # Design system
├── SSE_INTEGRATION.md   # SSE documentation
├── architecture.md      # System architecture
├── Gemfile              # Ruby dependencies
└── README.md            # This file
```

## Theme

This documentation uses [Just the Docs](https://just-the-docs.github.io/just-the-docs/) with a custom color scheme that matches the Nivo user-app styling (blue primary theme).

### Color Reference

| Color | Hex | Usage |
|:------|:----|:------|
| Primary | `#2563eb` | Links, buttons, accents |
| Primary Light | `#eff6ff` | Backgrounds |
| Primary Dark | `#1e40af` | Hover states |
| Success | `#22c55e` | Success indicators |
| Warning | `#f59e0b` | Warning indicators |
| Error | `#ef4444` | Error indicators |

## Adding New Pages

1. Create a new `.md` file in the docs root
2. Add Jekyll front matter:

```yaml
---
layout: default
title: Page Title
nav_order: 8
description: "Brief description"
permalink: /page-url
---

# Page Title

Content here...
```

3. Adjust `nav_order` to position in navigation

## GitHub Pages

This documentation is automatically built and deployed by GitHub Pages when changes are pushed to the main branch.

Configuration is in `_config.yml`:
- Custom domain: `docs.nivomoney.com`
- Remote theme: `just-the-docs/just-the-docs`

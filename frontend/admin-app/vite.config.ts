import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'
import tailwindcss from '@tailwindcss/vite'
import path from 'path'

// https://vite.dev/config/
export default defineConfig({
  plugins: [react(), tailwindcss()],
  server: {
    port: 3001,
    host: true,
    headers: {
      // Security headers for development
      'X-Frame-Options': 'DENY',
      'X-Content-Type-Options': 'nosniff',
      'Referrer-Policy': 'strict-origin-when-cross-origin',
    },
  },
  preview: {
    port: 3001,
    headers: {
      // Security headers for preview/testing
      'Strict-Transport-Security': 'max-age=31536000; includeSubDomains',
      'X-Frame-Options': 'DENY',
      'X-Content-Type-Options': 'nosniff',
      'Referrer-Policy': 'strict-origin-when-cross-origin',
      'Content-Security-Policy': "default-src 'self'; script-src 'self' 'unsafe-inline'; style-src 'self' 'unsafe-inline'; img-src 'self' data: https:; font-src 'self' data:; connect-src 'self' http://localhost:* ws://localhost:*",
      'Permissions-Policy': 'camera=(), microphone=(), geolocation=()',
    },
  },
  resolve: {
    alias: [
      // Local app alias
      { find: '@', replacement: path.resolve(__dirname, './src') },
      // Order matters: more specific paths first
      // Point to source files to ensure single React instance
      // (pre-built dist files cause hook context mismatch)
      { find: '@nivo/shared/components', replacement: path.resolve(__dirname, '../shared/components') },
      { find: '@nivo/shared/lib/utils', replacement: path.resolve(__dirname, '../shared/lib/utils.ts') },
      { find: '@nivo/shared', replacement: path.resolve(__dirname, '../shared/src') },
    ],
  },
  build: {
    outDir: 'dist',
    sourcemap: true,
  },
})

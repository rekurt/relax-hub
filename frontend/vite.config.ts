/// <reference types="vitest/config" />
import { defineConfig, loadEnv } from 'vite'
import react from '@vitejs/plugin-react'
import tailwindcss from '@tailwindcss/vite'
import path from 'path'

export default defineConfig(({ mode }) => {
  // Read repo-root .env (where BANI_SERVER_PORT lives) so the dev proxy stays in sync with the Go backend.
  const env = {
    ...loadEnv(mode, path.resolve(__dirname, '..'), ''),
    ...process.env,
  }
  const backendPort = env.BANI_SERVER_PORT || '8080'
  const backendUrl = env.BANI_BACKEND_URL || `http://localhost:${backendPort}`
  const backendWsUrl = backendUrl.replace(/^http/, 'ws')

  return {
    plugins: [tailwindcss(), react()],
    resolve: {
      alias: {
        '@': path.resolve(__dirname, './src'),
      },
    },
    test: {
      globals: true,
      environment: 'jsdom',
      setupFiles: './src/test-setup.ts',
      css: false,
      testTimeout: 30000,
      hookTimeout: 30000,
    },
    server: {
      host: true,
      allowedHosts: true,
      proxy: {
        '/api': {
          target: backendUrl,
          changeOrigin: true,
          ws: true,
        },
        '/ws': {
          target: backendWsUrl,
          ws: true,
        },
      },
    },
  }
})

import { defineConfig } from '@playwright/test'

const browserName = (process.env.E2E_BROWSER ?? 'chromium') as 'chromium' | 'firefox' | 'webkit'
const baseURL = process.env.E2E_BASE_URL ?? 'http://127.0.0.1:5173'

export default defineConfig({
  testDir: './tests/e2e',
  use: {
    baseURL,
    browserName,
  },
  reporter: [
    ['list'],
    ['json', { outputFile: 'tests/e2e/results/results.json' }],
  ],
  webServer: {
    command: 'npm run dev',
    url: baseURL,
    reuseExistingServer: true,
  },
})

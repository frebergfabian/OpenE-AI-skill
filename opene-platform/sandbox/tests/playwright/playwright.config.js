// Playwright mot OpenE-sandboxen. URL och konto kommer från sandboxctl (miljövariabler).
const { defineConfig, devices } = require('@playwright/test');

module.exports = defineConfig({
  testDir: './specs',
  timeout: 60_000,
  retries: 0,
  reporter: [['list'], ['html', { open: 'never', outputFolder: 'report' }]],
  outputDir: 'results',
  use: {
    baseURL: process.env.OPENE_URL || 'http://127.0.0.1:8480',
    ...(process.env.OPENE_BROWSER_PATH ? { launchOptions: { executablePath: process.env.OPENE_BROWSER_PATH } } : {}),
    locale: 'sv-SE',
    screenshot: 'only-on-failure',
    trace: 'retain-on-failure',
  },
  projects: [
    { name: 'desktop', use: { ...devices['Desktop Chrome'], viewport: { width: 1280, height: 900 } } },
    { name: 'mobil', use: { ...devices['Pixel 7'] } },
  ],
});

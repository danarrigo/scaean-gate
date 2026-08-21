import { defineConfig } from '@playwright/test';
import dotenv from 'dotenv';
import path from 'node:path';
dotenv.config({ path: path.resolve(__dirname, '../.env'), quiet: true });

export default defineConfig({
  testDir: './tests',
  timeout: 45_000,
  expect: { timeout: 10_000 },
  fullyParallel: false,
  workers: 1,
  use: { trace: 'retain-on-failure', screenshot: 'only-on-failure' },
  reporter: [['list']],
});

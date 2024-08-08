import { defineConfig } from '@playwright/test';
import * as path from 'path';
require('dotenv').config({ path: path.resolve(__dirname, '../.env-api') });

export default defineConfig({
  testDir: './tests',
  fullyParallel: true,
  forbidOnly: !!process.env.CI,
  retries: process.env.CI ? 2 : 0,
  workers: process.env.CI ? 1 : undefined,
  use: {
    // baseURL: 'https://issuer-node-api-v2-test.privado.id',//TODO get from container
    baseURL: `${process.env.ISSUER_API_URL}`,
    extraHTTPHeaders: {
      'Authorization': `Basic ${btoa(`${process.env.ISSUER_API_UI_AUTH_USER}:${process.env.ISSUER_API_UI_AUTH_PASSWORD}`)}`,
    }
  },
});

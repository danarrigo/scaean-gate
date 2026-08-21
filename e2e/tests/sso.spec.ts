import { expect, test } from '@playwright/test';

const password = process.env.SEED_USER_PASSWORD;
if (!password) throw new Error('SEED_USER_PASSWORD must be set in .env or the environment');

test('authorization code + PKCE creates independent sessions and SSO logout revokes both', async ({ page, context }) => {
  await page.goto('http://localhost:4201');
  await page.getByRole('button', { name: 'Continue with Scaean Gate' }).click();
  await expect(page).toHaveURL(/localhost:4200\/login/);
  await page.getByLabel('Email').fill('testuser@scaean-gate.com');
  await page.getByLabel('Password').fill(password);
  await page.getByRole('button', { name: 'Sign in' }).click();
  await expect(page).toHaveURL(/localhost:4201\/dashboard/);
  await expect(page.getByRole('heading', { name: /Hello, Test User/ })).toBeVisible();
  await expect(page.getByText('Local session created')).toBeVisible();

  const bolt = await context.newPage();
  await bolt.goto('http://localhost:4202');
  await bolt.getByRole('button', { name: 'Continue with Scaean Gate' }).click();
  await expect(bolt).toHaveURL(/localhost:4202\/dashboard/);
  await expect(bolt.getByRole('heading', { name: /Hello, Test User/ })).toBeVisible();

  page.on('dialog', dialog => dialog.accept());
  await page.getByRole('button', { name: 'End SSO session' }).click();
  await expect(page.getByRole('heading', { name: 'Your access was revoked' })).toBeVisible({ timeout: 25_000 });
  await expect(bolt.getByRole('heading', { name: 'Your access was revoked' })).toBeVisible({ timeout: 25_000 });
});

test('local logout leaves the central SSO session available', async ({ page }) => {
  await page.goto('http://localhost:4201');
  await page.getByRole('button', { name: 'Continue with Scaean Gate' }).click();
  await page.getByLabel('Email').fill('testuser@scaean-gate.com');
  await page.getByLabel('Password').fill(password);
  await page.getByRole('button', { name: 'Sign in' }).click();
  await expect(page.getByRole('heading', { name: /Hello, Test User/ })).toBeVisible();
  page.on('dialog', dialog => dialog.accept());
  await page.getByRole('button', { name: 'Local logout' }).click();
  await expect(page.getByRole('button', { name: 'Continue with Scaean Gate' })).toBeVisible();
  await page.getByRole('button', { name: 'Continue with Scaean Gate' }).click();
  await expect(page).toHaveURL(/localhost:4201\/dashboard/);
  await expect(page.getByRole('heading', { name: /Hello, Test User/ })).toBeVisible();
});

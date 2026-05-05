import { test, expect, Page } from '@playwright/test';
import { E2E_USER_EMAIL, E2E_USER_PASSWORD } from '../global-setup';

test.use({ storageState: { cookies: [], origins: [] } });

const TOPIC_URL = '/forum/topics/9999';

async function login(page: Page) {
  await page.goto('/login');
  await page.fill('#email', E2E_USER_EMAIL);
  await page.fill('#password', E2E_USER_PASSWORD);
  await page.click('button[type="submit"]');
  await page.waitForURL((url) => !url.pathname.includes('/login'), { timeout: 5000 });
}

// CHK-01: аватарка в nav header видна и открывает модалку
test('nav avatar visible and opens modal', async ({ page }) => {
  const consoleErrors: string[] = [];
  page.on('console', (msg) => {
    if (msg.type() === 'error') consoleErrors.push(msg.text());
  });

  await login(page);

  const navAvatar = page.locator('[data-testid="nav-avatar"]');
  await expect(navAvatar).toBeVisible();

  await Promise.all([
    page.waitForResponse((res) => res.url().includes('/modal') && res.status() < 400),
    navAvatar.click(),
  ]);

  await expect(page.locator('[data-testid="profile-modal"]')).toBeVisible();
  expect(consoleErrors).toHaveLength(0);
});

// CHK-04: пользователь без аватарки — инициалы в nav/forum/комментариях, 0 JS-ошибок
test('initials fallback, no console errors', async ({ page }) => {
  const consoleErrors: string[] = [];
  page.on('console', (msg) => {
    if (msg.type() === 'error') consoleErrors.push(msg.text());
  });

  await login(page);

  // Nav: e2e_user без аватарки → инициалы
  await page.goto('/');
  const navAvatar = page.locator('[data-testid="nav-avatar"]');
  await expect(navAvatar).toBeVisible();
  const navInitials = page.locator('[data-testid="nav-avatar-initials"]');
  await expect(navInitials).toBeVisible();

  // Forum: открыть тему — у поста e2e_user инициалы
  await page.goto(TOPIC_URL);
  await page.waitForSelector('#posts-list');
  const postAvatar = page.locator('[data-testid="post-avatar"]').first();
  await expect(postAvatar).toBeVisible();
  const postInitials = page.locator('[data-testid="post-avatar-initials"]').first();
  await expect(postInitials).toBeVisible();

  expect(consoleErrors).toHaveLength(0);
});

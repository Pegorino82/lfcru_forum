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

// CHK-02: аватарка автора видна у каждого поста в теме, клик открывает модалку
test('forum post avatar visible and opens modal', async ({ page }) => {
  await login(page);
  await page.goto(TOPIC_URL);
  await page.waitForSelector('#posts-list');

  // Проверяем первый пост с аватаркой (или инициалами)
  const postAvatar = page.locator('[data-testid="post-avatar"]').first();
  await expect(postAvatar).toBeVisible();

  await Promise.all([
    page.waitForResponse((res) => res.url().includes('/modal') && res.status() < 400),
    postAvatar.click(),
  ]);

  await expect(page.locator('[data-testid="profile-modal"]')).toBeVisible();
});

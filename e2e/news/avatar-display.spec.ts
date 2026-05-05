import { test, expect, Page } from '@playwright/test';
import { E2E_USER_EMAIL, E2E_USER_PASSWORD, E2E_NEWS_ID } from '../global-setup';

test.use({ storageState: { cookies: [], origins: [] } });

async function login(page: Page) {
  await page.goto('/login');
  await page.fill('#email', E2E_USER_EMAIL);
  await page.fill('#password', E2E_USER_PASSWORD);
  await page.click('button[type="submit"]');
  await page.waitForURL((url) => !url.pathname.includes('/login'), { timeout: 5000 });
}

// CHK-03: аватарка автора комментария видна в статье, клик открывает модалку
test('comment avatar visible and opens modal', async ({ page }) => {
  await login(page);
  await page.goto(`/news/${E2E_NEWS_ID}`);
  await page.waitForSelector('#comments-list');

  const commentAvatar = page.locator('[data-testid="comment-avatar"]').first();
  await expect(commentAvatar).toBeVisible();

  await Promise.all([
    page.waitForResponse((res) => res.url().includes('/modal') && res.status() < 400),
    commentAvatar.click(),
  ]);

  await expect(page.locator('[data-testid="profile-modal"]')).toBeVisible();
});

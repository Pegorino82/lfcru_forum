import { test, expect, Page } from '@playwright/test';
import { E2E_USER_EMAIL, E2E_USER_PASSWORD } from './global-setup';

const TOPIC_URL = '/forum/topics/9999';

async function login(page: Page) {
  await page.goto('/login');
  await page.fill('#email', E2E_USER_EMAIL);
  await page.fill('#password', E2E_USER_PASSWORD);
  await page.click('button[type="submit"]');
  // После логина HTMX делает редирект или перерисовку — ждём загрузки
  await page.waitForURL((url) => !url.pathname.includes('/login'), { timeout: 5000 });
}

test('ответ на комментарий содержит цитату и текст ответа', async ({ page }) => {
  await login(page);

  // Открыть тему
  await page.goto(TOPIC_URL);
  await page.waitForSelector('#posts-list');

  // Написать первое сообщение
  const originalText = 'E2E original post ' + Date.now();
  await page.fill('#post-content', originalText);

  await Promise.all([
    page.waitForResponse((res) =>
      res.url().includes('/forum/topics/9999/posts') && res.status() < 400,
    ),
    page.click('button[type="submit"]:has-text("Отправить")'),
  ]);

  // Дождаться появления поста в списке
  const firstPost = page.locator('#posts-list .post').first();
  await expect(firstPost).toBeVisible();

  // Нажать «Ответить» на первом посте
  await firstPost.locator('.reply-btn').click();

  // Форма ответа должна стать видимой (Alpine.js снимает x-cloak)
  const replyForm = firstPost.locator('.reply-form');
  await replyForm.waitFor({ state: 'visible' });

  // Убедиться, что в форме показана цитата
  await expect(replyForm.locator('.reply-form-quote')).toContainText('e2e_user');

  // Ввести текст ответа и отправить
  const replyText = 'E2E reply text ' + Date.now();
  await replyForm.locator('textarea[name="content"]').fill(replyText);

  await Promise.all([
    page.waitForResponse((res) =>
      res.url().includes('/forum/topics/9999/posts') && res.status() < 400,
    ),
    replyForm.locator('button[type="submit"]').click(),
  ]);

  // После HTMX outerHTML-swap #posts-list обновляется — ждём reply-пост
  const replyPost = page.locator('#posts-list .post.reply');
  await expect(replyPost).toBeVisible();

  // Цитата содержит текст оригинального поста
  await expect(replyPost.locator('.post-quote')).toContainText(originalText);

  // Текст ответа присутствует в контенте
  await expect(replyPost.locator('.post-content')).toContainText(replyText);
});

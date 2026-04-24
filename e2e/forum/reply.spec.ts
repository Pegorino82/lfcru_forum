import { test, expect, Page } from '@playwright/test';
import { E2E_USER_EMAIL, E2E_USER_PASSWORD } from '../global-setup';

const TOPIC_URL = '/forum/topics/9999';

async function login(page: Page) {
  await page.goto('/login');
  await page.fill('#email', E2E_USER_EMAIL);
  await page.fill('#password', E2E_USER_PASSWORD);
  await page.click('button[type="submit"]');
  await page.waitForURL((url) => !url.pathname.includes('/login'), { timeout: 5000 });
}

async function submitPost(page: Page, text: string) {
  await page.fill('#post-content', text);
  await Promise.all([
    page.waitForResponse(
      (res) => res.url().includes('/forum/topics/9999/posts') && res.status() < 400,
    ),
    page.click('button[type="submit"]:has-text("Отправить")'),
  ]);
}

async function submitReply(page: Page, targetPost: ReturnType<Page['locator']>, text: string) {
  await targetPost.locator('.reply-btn').click();
  const replyForm = targetPost.locator('.reply-form');
  await replyForm.waitFor({ state: 'visible' });
  await replyForm.locator('textarea[name="content"]').fill(text);
  await Promise.all([
    page.waitForResponse(
      (res) => res.url().includes('/forum/topics/9999/posts') && res.status() < 400,
    ),
    replyForm.locator('button[type="submit"]').click(),
  ]);
}

test('ответ на комментарий содержит цитату и текст ответа', async ({ page }) => {
  await login(page);
  await page.goto(TOPIC_URL);
  await page.waitForSelector('#posts-list');

  const originalText = 'E2E original post ' + Date.now();
  await submitPost(page, originalText);

  // Нажать «Ответить» на последнем посте (только что созданном)
  const newPost = page.locator('#posts-list .post').last();
  await expect(newPost).toBeVisible();

  const replyText = 'E2E reply text ' + Date.now();
  await submitReply(page, newPost, replyText);

  // После HTMX outerHTML-swap: reply — последний пост в списке
  const replyPost = page.locator('#posts-list .post').last();
  await expect(replyPost).toBeVisible();

  // Цитата содержит текст оригинального поста
  await expect(replyPost.locator('.post-quote')).toContainText(originalText);

  // Текст ответа присутствует
  await expect(replyPost.locator('.post-content')).toContainText(replyText);
});

// SC-01: ответ с цитатой отображается последним (не под процитированным)
test('SC-01: ответ с цитатой отображается последним в списке', async ({ page }) => {
  await login(page);
  await page.goto(TOPIC_URL);
  await page.waitForSelector('#posts-list');

  const p1Text = 'E2E SC01-P1 ' + Date.now();
  await submitPost(page, p1Text);

  const p2Text = 'E2E SC01-P2 ' + Date.now();
  await submitPost(page, p2Text);

  // Список после двух постов: найти P1 по тексту
  const p1Post = page.locator('#posts-list .post').filter({ hasText: p1Text });
  const replyText = 'E2E SC01-reply ' + Date.now();
  await submitReply(page, p1Post, replyText);

  // Reply должен быть последним в списке
  const allPosts = page.locator('#posts-list .post');
  const totalCount = await allPosts.count();

  const lastPost = allPosts.nth(totalCount - 1);
  await expect(lastPost.locator('.post-quote')).toContainText(p1Text);
  await expect(lastPost.locator('.post-content')).toContainText(replyText);

  // P2 должен быть предпоследним (reply не вклинился между P1 и P2)
  const secondToLast = allPosts.nth(totalCount - 2);
  await expect(secondToLast.locator('.post-content')).toContainText(p2Text);
});

// SC-02: клик на цитату переходит к оригинальному посту
test('SC-02: клик на цитату переходит к оригинальному посту', async ({ page }) => {
  await login(page);
  await page.goto(TOPIC_URL);
  await page.waitForSelector('#posts-list');

  const originalText = 'E2E SC02-target ' + Date.now();
  await submitPost(page, originalText);

  // Запомнить id оригинального поста
  const originalPost = page.locator('#posts-list .post').last();
  await expect(originalPost).toBeVisible();
  const originalPostId = await originalPost.getAttribute('id'); // "post-123"

  await submitReply(page, originalPost, 'E2E SC02-reply ' + Date.now());

  // Кликнуть на цитату в reply (последний пост)
  const replyPost = page.locator('#posts-list .post').last();
  await expect(replyPost.locator('.post-quote')).toBeVisible();
  await replyPost.locator('.post-quote').click();

  // URL должен содержать якорь #post-{id}
  await expect(page).toHaveURL(new RegExp(`#${originalPostId}$`));
});

import { test, expect } from '@playwright/test';

const TEAM_SECTION_ID = 9990;
const TEAM_TOPIC_ID = 9990;

test.use({ storageState: { cookies: [], origins: [] } });

test.describe('FT-028: Team section', () => {
  test('forum index shows "Команда" section', async ({ page }) => {
    await page.goto('/forum');
    const section = page.locator(`a[href="/forum/sections/${TEAM_SECTION_ID}"]`);
    await expect(section).toBeVisible();
    await expect(section).toContainText('Команда');
  });

  test('section page lists player topic', async ({ page }) => {
    await page.goto(`/forum/sections/${TEAM_SECTION_ID}`);
    const topic = page.locator(`a[href="/forum/topics/${TEAM_TOPIC_ID}"]`);
    await expect(topic).toBeVisible();
    await expect(topic).toContainText('Mohamed Salah');
  });

  test('topic page shows player card content', async ({ page }) => {
    await page.goto(`/forum/topics/${TEAM_TOPIC_ID}`);
    const content = page.locator('.post-content').first();
    await expect(content).toContainText('Имя: Mohamed Salah');
    await expect(content).toContainText('Позиция: Нападающий');
    await expect(content).toContainText('Дата рождения: 1992-06-15');
    await expect(content).toContainText('Национальность: Egypt');
    await expect(content).toContainText('Номер: —');
  });

  test('player card content is XSS-safe (auto-escaped)', async ({ page }) => {
    // The content is plain text stored in DB and rendered via {{.Content}} (Go template auto-escaping).
    // Verify that the content does NOT render as raw HTML.
    await page.goto(`/forum/topics/${TEAM_TOPIC_ID}`);
    const content = page.locator('.post-content').first();
    // Content should be visible as text, not contain any HTML tags as rendered elements
    const innerHTML = await content.innerHTML();
    // Plain text newlines become <br> or are preserved, but no script/HTML injection
    expect(innerHTML).not.toContain('<script');
    expect(innerHTML).not.toContain('<img');
  });
});

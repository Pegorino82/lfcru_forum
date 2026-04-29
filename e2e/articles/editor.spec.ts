import { test, expect, Page } from '@playwright/test';
import {
  E2E_ADMIN_EMAIL,
  E2E_ADMIN_PASSWORD,
  E2E_ARTICLE_ID,
} from '../global-setup';

test.use({ storageState: { cookies: [], origins: [] } });

const EDIT_URL = `/admin/articles/${E2E_ARTICLE_ID}/edit`;
const VIEW_URL = `/news/${E2E_ARTICLE_ID}`;

async function loginAdmin(page: Page) {
  await page.goto('/login');
  await page.fill('#email', E2E_ADMIN_EMAIL);
  await page.fill('#password', E2E_ADMIN_PASSWORD);
  await page.click('button[type="submit"]');
  await page.waitForURL((url) => !url.pathname.includes('/login'), { timeout: 5000 });
}

async function saveAndWait(page: Page) {
  await Promise.all([
    page.waitForURL((url) => url.pathname.includes('/edit'), { timeout: 10000 }),
    page.click('button[type="submit"].btn-primary'),
  ]);
}

async function publishArticle(page: Page) {
  const publishBtn = page.locator('.btn-success').first();
  if (await publishBtn.isVisible()) {
    await Promise.all([
      page.waitForURL((url) => url.pathname.includes('/edit'), { timeout: 5000 }),
      publishBtn.click(),
    ]);
  }
}

// CHK-01 / SC-01 / MET-01: HTML с форматированием сохраняется через hidden input
// и рендерится корректно на странице просмотра.
test('CHK-01: форматирование сохраняется и рендерится в публичном просмотре', async ({ page }) => {
  await loginAdmin(page);
  await page.goto(EDIT_URL);

  // Ждём инициализации TipTap
  await page.locator('.ProseMirror').waitFor({ state: 'visible', timeout: 10000 });

  // Устанавливаем контент напрямую в hidden input — минуем нестабильные TipTap keyboard interactions.
  // Это проверяет: bluemonday пропускает нужные теги, шаблон рендерит HTML без экранирования.
  const html =
    '<p><strong>Bold text</strong></p>' +
    '<p><em>Italic text</em></p>' +
    '<h2>Heading two</h2>' +
    '<p style="text-align:center">Centered text</p>';

  await page.evaluate((content) => {
    (document.getElementById('content-input') as HTMLInputElement).value = content;
  }, html);

  await saveAndWait(page);
  await publishArticle(page);

  await page.goto(VIEW_URL);
  const bodyHTML = await page.locator('.article-body').innerHTML();

  expect(bodyHTML).toContain('<strong>');
  expect(bodyHTML).toContain('<em>');
  expect(bodyHTML).toContain('<h2>');
  expect(bodyHTML).toMatch(/text-align:\s*center/);
});

// CHK-02 / SC-02: вставка изображения с подписью — figure/img/figcaption в DOM просмотра.
test('CHK-02: вставка изображения с подписью отображается в просмотре', async ({ page }) => {
  await loginAdmin(page);
  await page.goto(EDIT_URL);

  const editor = page.locator('.ProseMirror');
  await editor.waitFor({ state: 'visible', timeout: 10000 });

  // Регистрируем обработчик prompt ДО клика на кнопку
  page.on('dialog', async (dialog) => {
    await dialog.accept('E2E test caption');
  });

  const [fileChooser] = await Promise.all([
    page.waitForEvent('filechooser'),
    page.click('button[data-action="image-upload"]'),
  ]);

  // 1×1 PNG
  await fileChooser.setFiles({
    name: 'test.png',
    mimeType: 'image/png',
    buffer: Buffer.from(
      'iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mNk+M9QDwADhgGAWjR9awAAAABJRU5ErkJggg==',
      'base64',
    ),
  });

  // Ждём появления figure в редакторе
  await expect(editor.locator('figure')).toBeVisible({ timeout: 10000 });

  await saveAndWait(page);
  // Статья уже опубликована после CHK-01
  await page.goto(VIEW_URL);
  const bodyHTML = await page.locator('.article-body').innerHTML();

  expect(bodyHTML).toContain('<figure');
  expect(bodyHTML).toContain('<img');
  expect(bodyHTML).toContain('<figcaption');
});

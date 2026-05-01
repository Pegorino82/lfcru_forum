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

  // Устанавливаем контент через TipTap API — submit-handler вызывает editor.getHTML(),
  // поэтому менять только #content-input недостаточно: TipTap перезапишет его при сабмите.
  const html =
    '<p><strong>Bold text</strong></p>' +
    '<p><em>Italic text</em></p>' +
    '<h2>Heading two</h2>' +
    '<p style="text-align:center">Centered text</p>';

  await page.evaluate((content) => {
    (window as any)._tiptapEditor.commands.setContent(content);
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

test('FT-024: image editor bug fixes', async ({ page }) => {
  await loginAdmin(page);
  await page.goto(EDIT_URL);

  const editor = page.locator('.ProseMirror');
  await editor.waitFor({ state: 'visible', timeout: 10000 });
  await page.evaluate(() => (window as any)._tiptapEditor.commands.clearContent());

  // 1. Проверка на дублирование (Bug #2)
  const imagesList = page.locator('#images-list');
  const initialImageCount = await imagesList.locator('.image-thumb').count();

  page.on('dialog', async (dialog) => {
    expect(dialog.message()).toContain('Подпись к изображению');
    await dialog.accept('FT-024 fix test');
  });

  const [fileChooser] = await Promise.all([
    page.waitForEvent('filechooser'),
    page.click('button[data-action="image-upload"]'),
  ]);
  await fileChooser.setFiles({
    name: 'test-ft024.png',
    mimeType: 'image/png',
    buffer: Buffer.from('iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mP8/5+hHgAHggJ/PchI7wAAAABJRU5ErkJggg==', 'base64'),
  });

  await expect(editor.locator('figure')).toBeVisible({ timeout: 10000 });

  // Убедимся, что изображение НЕ добавилось в боковой список
  await expect(imagesList.locator('.image-thumb')).toHaveCount(initialImageCount);

  await saveAndWait(page);
  await page.goto(VIEW_URL);

  const body = page.locator('.article-body');
  await expect(body.locator('figure')).toHaveCount(1);

  // 2. Проверка удаления (Bug #1)
  await page.goto(EDIT_URL);
  await editor.waitFor({ state: 'visible', timeout: 10000 });

  await editor.locator('figure').click();
  await page.evaluate(() => (window as any)._tiptapEditor.commands.deleteSelection());

  await expect(editor.locator('figure')).toHaveCount(0);
  await saveAndWait(page);

  await page.goto(VIEW_URL);
  await expect(body.locator('figure')).toHaveCount(0);

  // 3. Проверка стиля (Bug #3 - full width)
  // Для этого вставим изображение еще раз
  await page.goto(EDIT_URL);
  await editor.waitFor({ state: 'visible', timeout: 10000 });
  const [fileChooser2] = await Promise.all([
    page.waitForEvent('filechooser'),
    page.click('button[data-action="image-upload"]'),
  ]);
  await fileChooser2.setFiles({
    name: 'test-ft024-2.png',
    mimeType: 'image/png',
    buffer: Buffer.from('iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAAADUlEQVR42mNkYPhfDwAChwGA60e6kgAAAABJRU5ErkJggg==', 'base64'),
  });

  await expect(editor.locator('figure')).toBeVisible({ timeout: 10000 });
  await saveAndWait(page);
  await page.goto(VIEW_URL);

  const image = body.locator('figure img');
  await expect(image).toBeVisible();
  const imageStyle = await image.evaluate(el => window.getComputedStyle(el));
  expect(imageStyle.maxWidth).toBe('100%');
});

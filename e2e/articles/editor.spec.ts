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

// CHK-01 / SC-01 / MET-01: форматирование bold/italic/h2/align-center
// сохраняется и отображается в публичном просмотре.
test('CHK-01: форматирование сохраняется и рендерится в публичном просмотре', async ({ page }) => {
  await loginAdmin(page);
  await page.goto(EDIT_URL);

  // Ждём инициализации TipTap
  const editor = page.locator('.ProseMirror');
  await editor.waitFor({ state: 'visible', timeout: 10000 });

  // Сбрасываем контент — кликаем, выделяем всё, удаляем
  await editor.click();
  await page.keyboard.press('Control+A');
  await page.keyboard.press('Backspace');

  // Вводим текст и применяем форматирование
  await page.keyboard.type('Bold text');
  await page.keyboard.press('Control+A');
  await page.click('button[data-action="bold"]');

  await editor.click();
  await page.keyboard.press('End');
  await page.keyboard.press('Enter');

  await page.click('button[data-action="italic"]');
  await page.keyboard.type('Italic text');
  await page.click('button[data-action="italic"]');
  await page.keyboard.press('Enter');

  await page.click('button[data-action="h2"]');
  await page.keyboard.type('Heading two');
  await page.click('button[data-action="h2"]');
  await page.keyboard.press('Enter');

  await page.click('button[data-action="align-center"]');
  await page.keyboard.type('Centered text');

  // Сохраняем
  await Promise.all([
    page.waitForURL((url) => url.pathname.includes('/edit'), { timeout: 10000 }),
    page.click('button[type="submit"].btn-primary'),
  ]);

  // Публикуем статью
  await Promise.all([
    page.waitForURL((url) => url.pathname.includes('/edit'), { timeout: 5000 }),
    page.click('.btn-success'),
  ]);

  // Проверяем публичный просмотр
  await page.goto(VIEW_URL);
  const bodyHTML = await page.locator('main').innerHTML();

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

  // Ставим обработчик dialog ПЕРЕД кликом кнопки (prompt появится во время обработки upload)
  page.on('dialog', async (dialog) => {
    await dialog.accept('E2E test caption');
  });

  // Кликаем кнопку image-upload — ожидаем file chooser
  const [fileChooser] = await Promise.all([
    page.waitForEvent('filechooser'),
    page.click('button[data-action="image-upload"]'),
  ]);

  // Загружаем тестовое изображение (1×1 PNG, base64)
  await fileChooser.setFiles({
    name: 'test.png',
    mimeType: 'image/png',
    buffer: Buffer.from(
      'iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mNk+M9QDwADhgGAWjR9awAAAABJRU5ErkJggg==',
      'base64',
    ),
  });

  // Ждём появления img в редакторе
  await expect(editor.locator('img')).toBeVisible({ timeout: 10000 });

  // Сохраняем
  await Promise.all([
    page.waitForURL((url) => url.pathname.includes('/edit'), { timeout: 10000 }),
    page.click('button[type="submit"].btn-primary'),
  ]);

  // Проверяем публичный просмотр (статья уже опубликована после CHK-01)
  await page.goto(VIEW_URL);
  const bodyHTML = await page.locator('main').innerHTML();

  expect(bodyHTML).toContain('<figure');
  expect(bodyHTML).toContain('<img');
  expect(bodyHTML).toContain('<figcaption');
});

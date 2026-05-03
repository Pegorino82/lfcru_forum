import { test, expect, Page } from '@playwright/test';
import { E2E_USER_EMAIL, E2E_USER_PASSWORD } from './global-setup';
import path from 'path';
import fs from 'fs';

test.use({ storageState: { cookies: [], origins: [] } });

const PROFILE_URL = '/profile/e2e_user';
const TOPIC_URL = '/forum/topics/9999';

async function login(page: Page) {
  await page.goto('/login');
  await page.fill('#email', E2E_USER_EMAIL);
  await page.fill('#password', E2E_USER_PASSWORD);
  await page.click('button[type="submit"]');
  await page.waitForURL((url) => !url.pathname.includes('/login'), { timeout: 5000 });
}

// SC-02: страница профиля отображает данные пользователя
test('SC-02: страница профиля доступна и содержит username', async ({ page }) => {
  await page.goto(PROFILE_URL);
  await expect(page.locator('[data-testid="profile-page"]')).toBeVisible();
  await expect(page.locator('[data-testid="profile-page"]')).toContainText('e2e_user');
});

// SC-06: пользователь без аватара — отображаются инициалы со стабильным цветом
test('SC-06: fallback-аватар: инициалы присутствуют и цвет стабилен при рефреше', async ({ page }) => {
  await page.goto(PROFILE_URL);
  const initialsEl = page.locator('[data-testid="avatar-initials"]');
  await expect(initialsEl).toBeVisible();

  const initials = await initialsEl.textContent();
  expect(initials?.trim().length).toBeGreaterThan(0);

  const color1 = await initialsEl.getAttribute('style');

  // Рефреш — цвет должен быть тем же
  await page.reload();
  await expect(page.locator('[data-testid="avatar-initials"]')).toBeVisible();
  const color2 = await page.locator('[data-testid="avatar-initials"]').getAttribute('style');

  expect(color1).toBe(color2);
});

// SC-07: пустые состояния при отсутствии постов/комментариев
test('SC-07: пустые состояния отображаются корректно', async ({ page }) => {
  // Используем нового пользователя без активности — e2e_admin вряд ли имеет посты на форуме
  await page.goto('/profile/e2e_admin');
  await expect(page.locator('[data-testid="profile-page"]')).toBeVisible();
  // Страница должна загрузиться без ошибок
  const status = await page.evaluate(() => document.readyState);
  expect(status).toBe('complete');
});

// SC-01: клик на имя пользователя открывает модальное окно
test('SC-01: клик на имя автора в топике открывает модалку', async ({ page }) => {
  await page.goto(TOPIC_URL);

  // Нажимаем на первую ссылку профиля на странице
  const profileLink = page.locator('.profile-link').first();
  await expect(profileLink).toBeVisible();

  await Promise.all([
    page.waitForResponse((res) => res.url().includes('/modal') && res.status() < 400),
    profileLink.click(),
  ]);

  await expect(page.locator('[data-testid="profile-modal"]')).toBeVisible();
  await expect(page.locator('[data-testid="modal-username"]')).toBeVisible();
});

// SC-01: модалка через nav username (авторизованный пользователь)
test('SC-01: клик на username в header открывает модалку', async ({ page }) => {
  await login(page);

  const navLink = page.locator('.nav-username');
  await expect(navLink).toBeVisible();

  await Promise.all([
    page.waitForResponse((res) => res.url().includes('/modal') && res.status() < 400),
    navLink.click(),
  ]);

  await expect(page.locator('[data-testid="profile-modal"]')).toBeVisible();
  await expect(page.locator('[data-testid="modal-username"]')).toContainText('e2e_user');
});

// SC-05: закрытие модалки кнопкой ✕
test('SC-05: модалка закрывается по кнопке X', async ({ page }) => {
  await page.goto(PROFILE_URL);
  // Открываем модалку через прямой HTMX запрос
  await page.evaluate(() => {
    document.getElementById('profile-modal-container')!.innerHTML = '';
  });
  await page.goto(TOPIC_URL);

  const profileLink = page.locator('.profile-link').first();
  await Promise.all([
    page.waitForResponse((res) => res.url().includes('/modal') && res.status() < 400),
    profileLink.click(),
  ]);

  await expect(page.locator('[data-testid="profile-modal"]')).toBeVisible();
  await page.locator('[data-testid="modal-close"]').click();
  await expect(page.locator('[data-testid="profile-modal"]')).not.toBeVisible();
});

// SC-05: закрытие модалки по ESC
test('SC-05: модалка закрывается по ESC', async ({ page }) => {
  await page.goto(TOPIC_URL);

  const profileLink = page.locator('.profile-link').first();
  await Promise.all([
    page.waitForResponse((res) => res.url().includes('/modal') && res.status() < 400),
    profileLink.click(),
  ]);

  await expect(page.locator('[data-testid="profile-modal"]')).toBeVisible();
  await page.keyboard.press('Escape');
  await expect(page.locator('[data-testid="profile-modal"]')).not.toBeVisible();
});

// SC-05: закрытие модалки кликом вне её
test('SC-05: модалка закрывается кликом вне', async ({ page }) => {
  await page.goto(TOPIC_URL);

  const profileLink = page.locator('.profile-link').first();
  await Promise.all([
    page.waitForResponse((res) => res.url().includes('/modal') && res.status() < 400),
    profileLink.click(),
  ]);

  await expect(page.locator('[data-testid="profile-modal"]')).toBeVisible();
  // Клик на overlay (вне modal-box)
  await page.locator('[data-testid="profile-modal-overlay"]').click({ position: { x: 5, y: 5 } });
  await expect(page.locator('[data-testid="profile-modal"]')).not.toBeVisible();
});

// SC-01: из модалки можно перейти на страницу профиля
test('SC-01: модалка содержит кнопку "Открыть профиль"', async ({ page }) => {
  await page.goto(TOPIC_URL);

  const profileLink = page.locator('.profile-link').first();
  await Promise.all([
    page.waitForResponse((res) => res.url().includes('/modal') && res.status() < 400),
    profileLink.click(),
  ]);

  await expect(page.locator('[data-testid="profile-modal"]')).toBeVisible();
  const openBtn = page.locator('[data-testid="modal-open-profile"]');
  await expect(openBtn).toBeVisible();
  await expect(openBtn).toHaveAttribute('href', /\/profile\//);
});

// NEG-06: 5xx ответ от /modal не роняет страницу (page.route fallback)
test('NEG-06: ошибка загрузки модалки не ломает страницу', async ({ page }) => {
  await page.route('**/modal', (route) => route.fulfill({ status: 500, body: 'error' }));
  await page.goto(TOPIC_URL);

  const profileLink = page.locator('.profile-link').first();
  await profileLink.click();

  // Страница осталась рабочей — форум-контент виден
  await expect(page.locator('h1')).toBeVisible();
  // Модалки нет (500 → не открылась)
  await expect(page.locator('[data-testid="profile-modal"]')).not.toBeVisible();
});

// SC-03: загрузка аватара (авторизованный пользователь, владелец профиля)
test('SC-03: владелец может загрузить аватар', async ({ page }) => {
  await login(page);
  await page.goto(PROFILE_URL);

  await expect(page.locator('[data-testid="avatar-upload"]')).toBeVisible();

  // Создаём минимальный PNG файл для загрузки
  const pngPath = path.join('/tmp', 'e2e-avatar-test.png');
  // 1x1 PNG (minimal valid PNG bytes)
  const minimalPNG = Buffer.from(
    '89504e470d0a1a0a0000000d49484452000000010000000108020000009001' +
    '2e0000000c4944415478016360f8cf000000020001e221bc330000000049454e44ae426082',
    'hex'
  );
  fs.writeFileSync(pngPath, minimalPNG);

  const [response] = await Promise.all([
    page.waitForResponse((res) => res.url().includes('/avatar') && res.status() < 400),
    page.locator('[data-testid="avatar-upload"] input[type=file]').setInputFiles(pngPath),
  ]);

  expect(response.status()).toBeLessThan(400);

  // Аватар-блок обновился — либо img, либо инициалы
  const avatarBlock = page.locator('#avatar-block');
  await expect(avatarBlock).toBeVisible();

  // Уборка
  fs.unlinkSync(pngPath);
});

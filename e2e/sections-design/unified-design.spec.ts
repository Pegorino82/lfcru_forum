import { test, expect } from '@playwright/test';

const NEWS_URL = '/news';
const FORUM_URL = '/forum';
const SECTION_URL = '/forum/sections/9999';

// --- SC-01: News cards with image/placeholder, title, excerpt, date, comments ---

test.describe('News list card layout (REQ-01, SC-01)', () => {
  test('news page renders card grid with expected elements', async ({ page }) => {
    await page.goto(NEWS_URL);
    await expect(page.locator('.news-list-title')).toBeVisible();

    const cards = page.locator('.news-card');
    const count = await cards.count();

    if (count > 0) {
      const firstCard = cards.first();
      // Each card has either an image or a placeholder
      const hasImage = await firstCard.locator('.news-hero').count();
      const hasPlaceholder = await firstCard.locator('.news-placeholder').count();
      expect(hasImage + hasPlaceholder).toBeGreaterThan(0);

      // Each card has title, meta
      await expect(firstCard.locator('.news-body h3 a')).toBeVisible();
      await expect(firstCard.locator('.news-meta')).toBeVisible();
    }
  });

  test('no JS console errors on /news', async ({ page }) => {
    const errors: string[] = [];
    page.on('console', (msg) => {
      if (msg.type() === 'error') errors.push(msg.text());
    });
    await page.goto(NEWS_URL);
    await page.waitForLoadState('networkidle');
    expect(errors).toEqual([]);
  });
});

// --- SC-02: Responsive grid columns ---

test.describe('News grid responsiveness (REQ-02, SC-02)', () => {
  test('1 column at 400px viewport', async ({ page }) => {
    await page.setViewportSize({ width: 400, height: 800 });
    await page.goto(NEWS_URL);

    const grid = page.locator('.news-grid');
    if (await grid.count() > 0) {
      const style = await grid.evaluate((el) => getComputedStyle(el).gridTemplateColumns);
      const columns = style.split(/\s+/).filter((s) => s !== '');
      expect(columns.length).toBe(1);
    }
  });

  test('2 columns at 800px viewport', async ({ page }) => {
    await page.setViewportSize({ width: 800, height: 800 });
    await page.goto(NEWS_URL);

    const grid = page.locator('.news-grid');
    if (await grid.count() > 0) {
      const style = await grid.evaluate((el) => getComputedStyle(el).gridTemplateColumns);
      const columns = style.split(/\s+/).filter((s) => s !== '');
      expect(columns.length).toBe(2);
    }
  });

  test('3 columns at 1280px viewport', async ({ page }) => {
    await page.setViewportSize({ width: 1280, height: 800 });
    await page.goto(NEWS_URL);

    const grid = page.locator('.news-grid');
    if (await grid.count() > 0) {
      const style = await grid.evaluate((el) => getComputedStyle(el).gridTemplateColumns);
      const columns = style.split(/\s+/).filter((s) => s !== '');
      expect(columns.length).toBe(3);
    }
  });
});

// --- SC-03, NEG-01, NEG-02: Pagination ---

test.describe('News pagination (REQ-03, SC-03)', () => {
  test('pagination nav is visible when multiple pages exist', async ({ page }) => {
    await page.goto(NEWS_URL);
    // Pagination only shown when TotalPages > 1, so check conditionally
    const pagination = page.locator('.pagination');
    const totalCards = await page.locator('.news-card').count();
    // If we see exactly 9 cards (page size), pagination likely exists
    if (totalCards >= 9) {
      await expect(pagination).toBeVisible();
    }
  });

  test('page=0 returns HTTP 200 (NEG-01)', async ({ page }) => {
    const response = await page.goto(`${NEWS_URL}?page=0`);
    expect(response?.status()).toBe(200);
  });

  test('page=-1 returns HTTP 200 (NEG-01)', async ({ page }) => {
    const response = await page.goto(`${NEWS_URL}?page=-1`);
    expect(response?.status()).toBe(200);
  });

  test('page=999 returns HTTP 200 (NEG-02)', async ({ page }) => {
    const response = await page.goto(`${NEWS_URL}?page=999`);
    expect(response?.status()).toBe(200);
  });
});

// --- SC-04: Forum sections full-width cards ---

test.describe('Forum sections layout (REQ-04, SC-04)', () => {
  test('sections render as full-width cards', async ({ page }) => {
    await page.setViewportSize({ width: 1280, height: 800 });
    await page.goto(FORUM_URL);

    const cards = page.locator('.section-card');
    const count = await cards.count();

    if (count > 0) {
      const firstCard = cards.first();
      await expect(firstCard.locator('h3 a')).toBeVisible();
      await expect(firstCard.locator('.topic-count')).toBeVisible();

      // Card should span full container width (minus padding)
      const containerWidth = await page.locator('.forum-index').evaluate((el) => el.clientWidth);
      const cardWidth = await firstCard.evaluate((el) => el.clientWidth);
      // Allow some tolerance for padding/margins
      expect(cardWidth).toBeGreaterThan(containerWidth * 0.9);
    }
  });

  test('no JS console errors on /forum', async ({ page }) => {
    const errors: string[] = [];
    page.on('console', (msg) => {
      if (msg.type() === 'error') errors.push(msg.text());
    });
    await page.goto(FORUM_URL);
    await page.waitForLoadState('networkidle');
    expect(errors).toEqual([]);
  });
});

// --- SC-05: Forum topics unified style ---

test.describe('Forum topics layout (REQ-05, SC-05)', () => {
  test('topics render with unified card style', async ({ page }) => {
    await page.goto(SECTION_URL);

    const rows = page.locator('.topic-row');
    const count = await rows.count();

    if (count > 0) {
      const firstRow = rows.first();
      await expect(firstRow.locator('.topic-info > a')).toBeVisible();
      await expect(firstRow.locator('.topic-stats')).toBeVisible();

      // Verify border-radius matches unified style (6px)
      const borderRadius = await firstRow.evaluate((el) => getComputedStyle(el).borderRadius);
      expect(borderRadius).toBe('6px');
    }
  });

  test('breadcrumbs are present', async ({ page }) => {
    await page.goto(SECTION_URL);
    const breadcrumbs = page.locator('.breadcrumbs');
    await expect(breadcrumbs).toBeVisible();
    await expect(breadcrumbs.locator('a[href="/forum"]')).toBeVisible();
  });

  test('no JS console errors on forum section', async ({ page }) => {
    const errors: string[] = [];
    page.on('console', (msg) => {
      if (msg.type() === 'error') errors.push(msg.text());
    });
    await page.goto(SECTION_URL);
    await page.waitForLoadState('networkidle');
    expect(errors).toEqual([]);
  });
});

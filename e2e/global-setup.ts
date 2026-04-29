import { Client } from 'pg';
import bcrypt from 'bcrypt';

const E2E_USER_ID = 9999;
const E2E_SECTION_ID = 9999;
const E2E_TOPIC_ID = 9999;

export const E2E_ADMIN_ID = 9998;
export const E2E_ARTICLE_ID = 9998;

export const E2E_USER_EMAIL = 'e2e@test.local';
export const E2E_USER_PASSWORD = 'e2e_pass123';
export const E2E_ADMIN_EMAIL = 'e2e_admin@test.local';
export const E2E_ADMIN_PASSWORD = 'e2e_admin_password';

export default async function globalSetup() {
  const client = new Client({
    host: process.env.PW_DB_HOST ?? 'localhost',
    port: 5432,
    user: 'postgres',
    password: 'postgres',
    database: 'lfcru_test',
  });

  await client.connect();

  try {
    const passHash = await bcrypt.hash(E2E_USER_PASSWORD, 4);
    const passHashBuf = Buffer.from(passHash);

    // Пользователь
    await client.query(
      `INSERT INTO users (id, username, email, pass_hash, role, is_active)
       OVERRIDING SYSTEM VALUE
       VALUES ($1, 'e2e_user', $2, $3, 'user', true)
       ON CONFLICT DO NOTHING`,
      [E2E_USER_ID, E2E_USER_EMAIL, passHashBuf],
    );

    // Секция форума
    await client.query(
      `INSERT INTO forum_sections (id, title, description, sort_order)
       OVERRIDING SYSTEM VALUE
       VALUES ($1, 'E2E Test Section', '', 999)
       ON CONFLICT DO NOTHING`,
      [E2E_SECTION_ID],
    );

    // Тема форума
    await client.query(
      `INSERT INTO forum_topics (id, section_id, title, author_id)
       OVERRIDING SYSTEM VALUE
       VALUES ($1, $2, 'E2E Test Topic', $3)
       ON CONFLICT DO NOTHING`,
      [E2E_TOPIC_ID, E2E_SECTION_ID, E2E_USER_ID],
    );

    // Администратор для E2E-тестов редактора
    const adminPassHash = await bcrypt.hash(E2E_ADMIN_PASSWORD, 4);
    const adminPassHashBuf = Buffer.from(adminPassHash);

    await client.query(
      `INSERT INTO users (id, username, email, pass_hash, role, is_active)
       OVERRIDING SYSTEM VALUE
       VALUES ($1, 'e2e_admin', $2, $3, 'admin', true)
       ON CONFLICT DO NOTHING`,
      [E2E_ADMIN_ID, E2E_ADMIN_EMAIL, adminPassHashBuf],
    );

    // Тестовая статья для E2E-тестов редактора (черновик, автор — e2e_admin)
    await client.query(
      `INSERT INTO news (id, title, content, status, author_id)
       OVERRIDING SYSTEM VALUE
       VALUES ($1, 'E2E Test Article', '<p>E2E test article body</p>', 'draft', $2)
       ON CONFLICT DO NOTHING`,
      [E2E_ARTICLE_ID, E2E_ADMIN_ID],
    );
  } finally {
    await client.end();
  }
}

import { Client } from 'pg';

const E2E_USER_ID = 9999;
const E2E_SECTION_ID = 9999;
const E2E_TOPIC_ID = 9999;
const E2E_ADMIN_ID = 9998;
const E2E_ARTICLE_ID = 9998;

export default async function globalTeardown() {
  const client = new Client({
    host: process.env.PW_DB_HOST ?? 'localhost',
    port: 5432,
    user: 'postgres',
    password: 'postgres',
    database: 'lfcru_test',
  });

  await client.connect();

  try {
    // Посты удаляются каскадно вместе с темой
    await client.query('DELETE FROM forum_topics WHERE id = $1', [E2E_TOPIC_ID]);
    await client.query('DELETE FROM forum_sections WHERE id = $1', [E2E_SECTION_ID]);
    await client.query('DELETE FROM users WHERE id = $1', [E2E_USER_ID]);
    // E2E article images удаляются каскадно вместе со статьёй
    await client.query('DELETE FROM news WHERE id = $1', [E2E_ARTICLE_ID]);
    await client.query('DELETE FROM users WHERE id = $1', [E2E_ADMIN_ID]);
  } finally {
    await client.end();
  }
}

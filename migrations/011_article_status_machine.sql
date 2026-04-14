-- +goose Up

CREATE TYPE news_status AS ENUM ('draft', 'in_review', 'published');

DROP INDEX IF EXISTS idx_news_published;

ALTER TABLE news
  ADD COLUMN status      news_status NOT NULL DEFAULT 'draft',
  ADD COLUMN reviewer_id BIGINT REFERENCES users(id);

UPDATE news SET status = (CASE WHEN is_published THEN 'published' ELSE 'draft' END)::news_status;

ALTER TABLE news DROP CONSTRAINT chk_news_published;
ALTER TABLE news DROP COLUMN is_published;

CREATE INDEX idx_news_published ON news (published_at DESC) WHERE status = 'published';

-- +goose Down

DROP INDEX IF EXISTS idx_news_published;
ALTER TABLE news ADD COLUMN is_published BOOLEAN NOT NULL DEFAULT false;
UPDATE news SET is_published = (status = 'published');
ALTER TABLE news ADD CONSTRAINT chk_news_published CHECK (is_published = false OR published_at IS NOT NULL);
ALTER TABLE news DROP COLUMN reviewer_id;
ALTER TABLE news DROP COLUMN status;
CREATE INDEX idx_news_published ON news (published_at DESC) WHERE is_published = true;
DROP TYPE IF EXISTS news_status;

-- +goose Up
ALTER TABLE forum_posts
ADD COLUMN parent_id BIGINT REFERENCES forum_posts(id) ON DELETE SET NULL,
ADD COLUMN parent_author_snapshot TEXT,
ADD COLUMN parent_content_snapshot TEXT;

-- Индекс для быстрого поиска по parent_id
CREATE INDEX idx_forum_posts_parent ON forum_posts (parent_id);

-- +goose Down
DROP INDEX IF EXISTS idx_forum_posts_parent;
ALTER TABLE forum_posts
DROP COLUMN parent_id,
DROP COLUMN parent_author_snapshot,
DROP COLUMN parent_content_snapshot;

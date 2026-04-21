-- +goose NO TRANSACTION
-- +goose Up
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_posts_topic_id_id ON forum_posts(topic_id, id);

-- +goose Down
DROP INDEX CONCURRENTLY IF EXISTS idx_posts_topic_id_id;

-- +goose Up
ALTER TABLE forum_sections
ADD COLUMN description TEXT NOT NULL DEFAULT '',
ADD COLUMN topic_count INT NOT NULL DEFAULT 0;

-- Обновить topic_count для существующих разделов
UPDATE forum_sections SET topic_count = (
    SELECT COUNT(*) FROM forum_topics WHERE section_id = forum_sections.id
);

-- Функция: обновление topic_count в forum_sections при INSERT/DELETE тем
-- +goose StatementBegin
CREATE OR REPLACE FUNCTION trg_forum_topics_update_section_count()
RETURNS TRIGGER AS $$
BEGIN
    IF TG_OP = 'INSERT' THEN
        UPDATE forum_sections
        SET topic_count = topic_count + 1
        WHERE id = NEW.section_id;
    ELSIF TG_OP = 'DELETE' THEN
        UPDATE forum_sections
        SET topic_count = GREATEST(topic_count - 1, 0)
        WHERE id = OLD.section_id;
    END IF;
    RETURN NULL;
END;
$$ LANGUAGE plpgsql;
-- +goose StatementEnd

CREATE TRIGGER trg_forum_topics_update_section_count
AFTER INSERT OR DELETE ON forum_topics
FOR EACH ROW EXECUTE FUNCTION trg_forum_topics_update_section_count();

-- Индекс для оптимизации запросов
CREATE INDEX idx_forum_topics_section ON forum_topics (section_id);

-- +goose Down
DROP TRIGGER IF EXISTS trg_forum_topics_update_section_count ON forum_topics;
DROP FUNCTION IF EXISTS trg_forum_topics_update_section_count;
DROP INDEX IF EXISTS idx_forum_topics_section;
ALTER TABLE forum_sections
DROP COLUMN topic_count,
DROP COLUMN description;

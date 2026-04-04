-- +goose Up
CREATE TABLE matches (
    id           BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    opponent     VARCHAR(255)  NOT NULL,
    match_date   TIMESTAMPTZ   NOT NULL,
    tournament   VARCHAR(255)  NOT NULL,
    is_home      BOOLEAN       NOT NULL DEFAULT true,
    created_at   TIMESTAMPTZ   NOT NULL DEFAULT now(),
    updated_at   TIMESTAMPTZ   NOT NULL DEFAULT now()
);

CREATE INDEX idx_matches_date ON matches (match_date ASC);

CREATE TABLE forum_sections (
    id          BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    title       VARCHAR(255) NOT NULL,
    sort_order  INT          NOT NULL DEFAULT 0,
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT now()
);

CREATE TABLE forum_topics (
    id             BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    section_id     BIGINT        NOT NULL REFERENCES forum_sections(id) ON DELETE RESTRICT,
    title          VARCHAR(255)  NOT NULL,
    author_id      BIGINT        NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    post_count     INT           NOT NULL DEFAULT 0,
    last_post_at   TIMESTAMPTZ,
    last_post_by   BIGINT        REFERENCES users(id) ON DELETE SET NULL,
    created_at     TIMESTAMPTZ   NOT NULL DEFAULT now(),
    updated_at     TIMESTAMPTZ   NOT NULL DEFAULT now()
);

CREATE INDEX idx_forum_topics_last_post ON forum_topics (last_post_at DESC NULLS LAST);

CREATE TABLE forum_posts (
    id          BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    topic_id    BIGINT       NOT NULL REFERENCES forum_topics(id) ON DELETE CASCADE,
    author_id   BIGINT       NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    content     TEXT         NOT NULL,
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ  NOT NULL DEFAULT now()
);

CREATE INDEX idx_forum_posts_topic ON forum_posts (topic_id, created_at ASC);

-- Триггер: обновление денормализованных полей forum_topics при INSERT/DELETE в forum_posts
-- +goose StatementBegin
CREATE OR REPLACE FUNCTION forum_topics_update_last_post()
RETURNS TRIGGER AS $$
BEGIN
    IF TG_OP = 'INSERT' THEN
        UPDATE forum_topics
        SET post_count  = post_count + 1,
            last_post_at = NEW.created_at,
            last_post_by = NEW.author_id,
            updated_at   = now()
        WHERE id = NEW.topic_id;
    ELSIF TG_OP = 'DELETE' THEN
        UPDATE forum_topics
        SET post_count   = GREATEST(post_count - 1, 0),
            last_post_at = (SELECT MAX(created_at) FROM forum_posts WHERE topic_id = OLD.topic_id AND id != OLD.id),
            last_post_by = (SELECT author_id FROM forum_posts WHERE topic_id = OLD.topic_id AND id != OLD.id ORDER BY created_at DESC LIMIT 1),
            updated_at   = now()
        WHERE id = OLD.topic_id;
    END IF;
    RETURN NULL;
END;
$$ LANGUAGE plpgsql;
-- +goose StatementEnd

CREATE TRIGGER trg_forum_posts_update_topic
AFTER INSERT OR DELETE ON forum_posts
FOR EACH ROW EXECUTE FUNCTION forum_topics_update_last_post();

-- +goose Down
DROP TRIGGER IF EXISTS trg_forum_posts_update_topic ON forum_posts;
DROP FUNCTION IF EXISTS forum_topics_update_last_post;
DROP TABLE IF EXISTS forum_posts;
DROP TABLE IF EXISTS forum_topics;
DROP TABLE IF EXISTS forum_sections;
DROP TABLE IF EXISTS matches;

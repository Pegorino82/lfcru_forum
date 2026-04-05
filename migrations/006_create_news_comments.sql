-- +goose Up
CREATE TABLE news_comments (
    id                      BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    news_id                 BIGINT       NOT NULL REFERENCES news(id) ON DELETE CASCADE,
    author_id               BIGINT       NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    parent_id               BIGINT       REFERENCES news_comments(id) ON DELETE SET NULL,
    parent_author_snapshot  TEXT,
    parent_content_snapshot TEXT,
    content                 TEXT         NOT NULL,
    created_at              TIMESTAMPTZ  NOT NULL DEFAULT now()
);

CREATE INDEX idx_news_comments_news ON news_comments (news_id, created_at ASC);

-- +goose Down
DROP TABLE IF EXISTS news_comments;

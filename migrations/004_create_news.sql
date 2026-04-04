-- +goose Up
CREATE TABLE news (
    id           BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    title        VARCHAR(255)  NOT NULL,
    content      TEXT          NOT NULL DEFAULT '',
    is_published BOOLEAN       NOT NULL DEFAULT false,
    author_id    BIGINT        NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    published_at TIMESTAMPTZ,
    CONSTRAINT chk_news_published CHECK (is_published = false OR published_at IS NOT NULL),
    created_at   TIMESTAMPTZ   NOT NULL DEFAULT now(),
    updated_at   TIMESTAMPTZ   NOT NULL DEFAULT now()
);

CREATE INDEX idx_news_published ON news (published_at DESC)
    WHERE is_published = true;

-- +goose Down
DROP TABLE IF EXISTS news;

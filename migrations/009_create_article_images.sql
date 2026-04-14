-- +goose Up
CREATE TABLE article_images (
    id                BIGSERIAL PRIMARY KEY,
    article_id        BIGINT NOT NULL REFERENCES news(id) ON DELETE CASCADE,
    filename          TEXT NOT NULL,
    original_filename TEXT NOT NULL,
    created_at        TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_article_images_article_id ON article_images(article_id);

-- +goose Down
DROP TABLE IF EXISTS article_images;

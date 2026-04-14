package admin

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// ArticleImage represents a stored image attached to a news article.
type ArticleImage struct {
	ID               int64
	ArticleID        int64
	Filename         string
	OriginalFilename string
	CreatedAt        time.Time
}

// ImagesRepo handles database operations for article_images.
type ImagesRepo struct {
	pool *pgxpool.Pool
}

// NewImagesRepo creates a new ImagesRepo.
func NewImagesRepo(pool *pgxpool.Pool) *ImagesRepo {
	return &ImagesRepo{pool: pool}
}

// Create inserts a new article_images record and sets img.ID and img.CreatedAt.
func (r *ImagesRepo) Create(ctx context.Context, img *ArticleImage) error {
	return r.pool.QueryRow(ctx, `
		INSERT INTO article_images (article_id, filename, original_filename)
		VALUES ($1, $2, $3)
		RETURNING id, created_at
	`, img.ArticleID, img.Filename, img.OriginalFilename).Scan(&img.ID, &img.CreatedAt)
}

// ListByArticleID returns all images for a given article, ordered by created_at ASC.
func (r *ImagesRepo) ListByArticleID(ctx context.Context, articleID int64) ([]ArticleImage, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, article_id, filename, original_filename, created_at
		FROM article_images
		WHERE article_id = $1
		ORDER BY created_at ASC
	`, articleID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := []ArticleImage{}
	for rows.Next() {
		var img ArticleImage
		if err := rows.Scan(&img.ID, &img.ArticleID, &img.Filename, &img.OriginalFilename, &img.CreatedAt); err != nil {
			return nil, err
		}
		result = append(result, img)
	}
	return result, rows.Err()
}

// GetByID returns the image with the given id, or nil if not found.
func (r *ImagesRepo) GetByID(ctx context.Context, id int64) (*ArticleImage, error) {
	img := &ArticleImage{}
	err := r.pool.QueryRow(ctx, `
		SELECT id, article_id, filename, original_filename, created_at
		FROM article_images
		WHERE id = $1
	`, id).Scan(&img.ID, &img.ArticleID, &img.Filename, &img.OriginalFilename, &img.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return img, nil
}

// Delete removes the article_images record with the given id.
func (r *ImagesRepo) Delete(ctx context.Context, id int64) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM article_images WHERE id = $1`, id)
	return err
}

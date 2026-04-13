package news

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repo struct{ pool *pgxpool.Pool }

func NewRepo(pool *pgxpool.Pool) *Repo { return &Repo{pool: pool} }

// LatestPublished возвращает до limit опубликованных новостей,
// отсортированных по published_at DESC. При отсутствии данных — пустой слайс.
func (r *Repo) LatestPublished(ctx context.Context, limit int) ([]News, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, title, published_at
		FROM news
		WHERE is_published = true
		ORDER BY published_at DESC
		LIMIT $1
	`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []News
	for rows.Next() {
		var n News
		if err := rows.Scan(&n.ID, &n.Title, &n.PublishedAt); err != nil {
			return nil, err
		}
		result = append(result, n)
	}
	if result == nil {
		result = []News{}
	}
	return result, rows.Err()
}

// ListPublished возвращает опубликованные новости с пагинацией и общее количество.
// Результаты отсортированы по published_at DESC.
func (r *Repo) ListPublished(ctx context.Context, limit, offset int) ([]News, int64, error) {
	var total int64
	err := r.pool.QueryRow(ctx, `
		SELECT COUNT(*) FROM news WHERE is_published = true
	`).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	rows, err := r.pool.Query(ctx, `
		SELECT id, title, published_at
		FROM news
		WHERE is_published = true
		ORDER BY published_at DESC
		LIMIT $1 OFFSET $2
	`, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	result := []News{}
	for rows.Next() {
		var n News
		if err := rows.Scan(&n.ID, &n.Title, &n.PublishedAt); err != nil {
			return nil, 0, err
		}
		result = append(result, n)
	}
	return result, total, rows.Err()
}

// GetPublishedByID возвращает опубликованную статью по ID.
// Если статья не найдена или не опубликована — возвращает nil, nil.
func (r *Repo) GetPublishedByID(ctx context.Context, id int64) (*News, error) {
	n := &News{}
	err := r.pool.QueryRow(ctx, `
		SELECT id, title, content, is_published, author_id, published_at, created_at, updated_at
		FROM news
		WHERE id = $1 AND is_published = true
	`, id).Scan(&n.ID, &n.Title, &n.Content, &n.IsPublished, &n.AuthorID,
		&n.PublishedAt, &n.CreatedAt, &n.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return n, nil
}

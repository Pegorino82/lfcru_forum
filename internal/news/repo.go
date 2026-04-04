package news

import (
	"context"

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

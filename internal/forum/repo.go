package forum

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Repo struct{ pool *pgxpool.Pool }

func NewRepo(pool *pgxpool.Pool) *Repo { return &Repo{pool: pool} }

// LatestActive возвращает до limit тем с последней активностью.
// Темы без сообщений (last_post_at IS NULL) не включаются.
func (r *Repo) LatestActive(ctx context.Context, limit int) ([]TopicWithLastAuthor, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT t.id, t.title, t.last_post_at, COALESCE(u.username, '[удалён]') AS last_post_by_name
		FROM forum_topics t
		LEFT JOIN users u ON u.id = t.last_post_by
		WHERE t.last_post_at IS NOT NULL
		ORDER BY t.last_post_at DESC
		LIMIT $1
	`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []TopicWithLastAuthor
	for rows.Next() {
		var t TopicWithLastAuthor
		if err := rows.Scan(&t.ID, &t.Title, &t.LastPostAt, &t.LastPostByName); err != nil {
			return nil, err
		}
		result = append(result, t)
	}
	if result == nil {
		result = []TopicWithLastAuthor{}
	}
	return result, rows.Err()
}

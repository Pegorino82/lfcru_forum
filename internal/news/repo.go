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
		WHERE status = 'published'
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
		SELECT COUNT(*) FROM news WHERE status = 'published'
	`).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	rows, err := r.pool.Query(ctx, `
		SELECT id, title, published_at
		FROM news
		WHERE status = 'published'
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

// ListImagesByArticleID returns images attached to an article, ordered by created_at ASC.
func (r *Repo) ListImagesByArticleID(ctx context.Context, articleID int64) ([]ImageView, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, filename
		FROM article_images
		WHERE article_id = $1
		ORDER BY created_at ASC
	`, articleID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := []ImageView{}
	for rows.Next() {
		var img ImageView
		if err := rows.Scan(&img.ID, &img.Filename); err != nil {
			return nil, err
		}
		result = append(result, img)
	}
	return result, rows.Err()
}

// GetPublishedByID возвращает опубликованную статью по ID.
// Если статья не найдена или не опубликована — возвращает nil, nil.
func (r *Repo) GetPublishedByID(ctx context.Context, id int64) (*News, error) {
	n := &News{}
	err := r.pool.QueryRow(ctx, `
		SELECT id, title, content, status, reviewer_id, author_id, published_at, created_at, updated_at
		FROM news
		WHERE id = $1 AND status = 'published'
	`, id).Scan(&n.ID, &n.Title, &n.Content, &n.Status, &n.ReviewerID, &n.AuthorID,
		&n.PublishedAt, &n.CreatedAt, &n.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return n, nil
}

// GetByIDAdmin возвращает статью по ID без фильтрации по статусу (для admin-панели).
// Если статья не найдена — возвращает nil, nil.
func (r *Repo) GetByIDAdmin(ctx context.Context, id int64) (*News, error) {
	n := &News{}
	err := r.pool.QueryRow(ctx, `
		SELECT id, title, content, status, reviewer_id, author_id, published_at, created_at, updated_at
		FROM news
		WHERE id = $1
	`, id).Scan(&n.ID, &n.Title, &n.Content, &n.Status, &n.ReviewerID, &n.AuthorID,
		&n.PublishedAt, &n.CreatedAt, &n.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return n, nil
}

// CreateDraft создаёт новую статью со статусом draft и заполняет n.ID, n.CreatedAt, n.UpdatedAt.
func (r *Repo) CreateDraft(ctx context.Context, n *News) error {
	return r.pool.QueryRow(ctx, `
		INSERT INTO news (title, content, author_id, status)
		VALUES ($1, $2, $3, 'draft')
		RETURNING id, created_at, updated_at
	`, n.Title, n.Content, n.AuthorID).Scan(&n.ID, &n.CreatedAt, &n.UpdatedAt)
}

// UpdateArticle обновляет заголовок и содержимое статьи.
func (r *Repo) UpdateArticle(ctx context.Context, n *News) error {
	tag, err := r.pool.Exec(ctx, `
		UPDATE news SET title = $2, content = $3, updated_at = now()
		WHERE id = $1
	`, n.ID, n.Title, n.Content)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}

// ChangeStatus меняет статус статьи. Если новый статус — 'published' и published_at ещё не задан,
// устанавливается текущее время.
func (r *Repo) ChangeStatus(ctx context.Context, id int64, status ArticleStatus, reviewerID *int64) error {
	tag, err := r.pool.Exec(ctx, `
		UPDATE news
		SET status = $2::news_status,
		    reviewer_id = $3,
		    published_at = CASE
		        WHEN $2::news_status = 'published' AND published_at IS NULL THEN now()
		        ELSE published_at
		    END,
		    updated_at = now()
		WHERE id = $1
	`, id, string(status), reviewerID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}

// ListByStatus возвращает все статьи с фильтрацией по статусу.
// Если status == "" — возвращает все статьи.
func (r *Repo) ListByStatus(ctx context.Context, status ArticleStatus) ([]News, error) {
	var rows pgx.Rows
	var err error

	if status == "" {
		rows, err = r.pool.Query(ctx, `
			SELECT id, title, status, author_id, published_at, created_at
			FROM news
			ORDER BY created_at DESC
		`)
	} else {
		rows, err = r.pool.Query(ctx, `
			SELECT id, title, status, author_id, published_at, created_at
			FROM news
			WHERE status = $1::news_status
			ORDER BY created_at DESC
		`, string(status))
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := []News{}
	for rows.Next() {
		var n News
		if err := rows.Scan(&n.ID, &n.Title, &n.Status, &n.AuthorID, &n.PublishedAt, &n.CreatedAt); err != nil {
			return nil, err
		}
		result = append(result, n)
	}
	return result, rows.Err()
}

package comment

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repo struct{ pool *pgxpool.Pool }

func NewRepo(pool *pgxpool.Pool) *Repo { return &Repo{pool: pool} }

// ListByNewsID возвращает все комментарии к статье, отсортированные хронологически.
// При отсутствии данных — пустой слайс.
func (r *Repo) ListByNewsID(ctx context.Context, newsID int64) ([]CommentView, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT
			c.id,
			c.news_id,
			c.author_id,
			u.username AS author_username,
			c.parent_id,
			c.parent_author_snapshot,
			c.parent_content_snapshot,
			c.content,
			c.created_at
		FROM news_comments c
		JOIN users u ON u.id = c.author_id
		WHERE c.news_id = $1
		ORDER BY c.created_at ASC, c.id ASC
		LIMIT 500
	`, newsID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []CommentView
	for rows.Next() {
		var cv CommentView
		if err := rows.Scan(
			&cv.ID, &cv.NewsID, &cv.AuthorID, &cv.AuthorUsername,
			&cv.ParentID, &cv.ParentAuthor, &cv.ParentSnippet,
			&cv.Content, &cv.CreatedAt,
		); err != nil {
			return nil, err
		}
		// Добавить троеточие к snippet если он был обрезан (ровно 100 рун)
		if cv.ParentSnippet != nil {
			runes := []rune(*cv.ParentSnippet)
			if len(runes) == 100 {
				s := *cv.ParentSnippet + "…"
				cv.ParentSnippet = &s
			}
		}
		result = append(result, cv)
	}
	if result == nil {
		result = []CommentView{}
	}
	return result, rows.Err()
}

// Create сохраняет комментарий в БД.
// Если ParentID != nil, выполняется в транзакции:
// проверяет существование родителя, ограничение глубины (≤ 1), заполняет snapshot-поля.
func (r *Repo) Create(ctx context.Context, c *Comment) (int64, error) {
	if c.ParentID == nil {
		return r.insertComment(ctx, c)
	}

	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback(ctx)

	// Шаг 1: получить данные родительского комментария
	var parentAuthor string
	var parentSnippet string
	var parentParentID *int64
	err = tx.QueryRow(ctx, `
		SELECT u.username, LEFT(nc.content, 100), nc.parent_id
		FROM news_comments nc
		JOIN users u ON u.id = nc.author_id
		WHERE nc.id = $1 AND nc.news_id = $2
	`, *c.ParentID, c.NewsID).Scan(&parentAuthor, &parentSnippet, &parentParentID)
	if errors.Is(err, pgx.ErrNoRows) {
		return 0, ErrParentNotFound
	}
	if err != nil {
		return 0, err
	}

	// Ограничение глубины: ответ на ответ запрещён
	if parentParentID != nil {
		return 0, ErrParentNotFound
	}

	c.ParentAuthorSnapshot = &parentAuthor
	c.ParentContentSnapshot = &parentSnippet

	// Шаг 2: вставка
	var id int64
	err = tx.QueryRow(ctx, `
		INSERT INTO news_comments
			(news_id, author_id, parent_id, parent_author_snapshot, parent_content_snapshot, content)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id
	`, c.NewsID, c.AuthorID, c.ParentID, c.ParentAuthorSnapshot, c.ParentContentSnapshot, c.Content).Scan(&id)
	if err != nil {
		return 0, mapFKViolation(err)
	}

	if err := tx.Commit(ctx); err != nil {
		return 0, err
	}
	return id, nil
}

func (r *Repo) insertComment(ctx context.Context, c *Comment) (int64, error) {
	var id int64
	err := r.pool.QueryRow(ctx, `
		INSERT INTO news_comments (news_id, author_id, content)
		VALUES ($1, $2, $3)
		RETURNING id
	`, c.NewsID, c.AuthorID, c.Content).Scan(&id)
	if err != nil {
		return 0, err
	}
	return id, nil
}

func mapFKViolation(err error) error {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23503" {
		return ErrParentNotFound
	}
	return err
}

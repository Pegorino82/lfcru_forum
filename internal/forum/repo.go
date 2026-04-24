package forum

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repo struct{ pool *pgxpool.Pool }

func NewRepo(pool *pgxpool.Pool) *Repo { return &Repo{pool: pool} }

// ListSections возвращает все разделы, отсортированные по sort_order, id.
func (r *Repo) ListSections(ctx context.Context) ([]SectionView, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, title, description, topic_count
		FROM forum_sections
		ORDER BY sort_order ASC, id ASC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []SectionView
	for rows.Next() {
		var s SectionView
		if err := rows.Scan(&s.ID, &s.Title, &s.Description, &s.TopicCount); err != nil {
			return nil, err
		}
		result = append(result, s)
	}
	if result == nil {
		result = []SectionView{}
	}
	return result, rows.Err()
}

// GetSection возвращает раздел по ID, или (nil, nil) если не найден.
func (r *Repo) GetSection(ctx context.Context, id int64) (*Section, error) {
	var s Section
	err := r.pool.QueryRow(ctx, `
		SELECT id, title, description, sort_order, topic_count, created_at
		FROM forum_sections
		WHERE id = $1
	`, id).Scan(&s.ID, &s.Title, &s.Description, &s.SortOrder, &s.TopicCount, &s.CreatedAt)

	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &s, nil
}

// ListTopicsBySection возвращает темы раздела, отсортированные по last_post_at DESC NULLS LAST.
func (r *Repo) ListTopicsBySection(ctx context.Context, sectionID int64) ([]TopicView, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT t.id, t.title, t.author_id, COALESCE(u.username, '[удалён]') AS author_username,
		       t.post_count, t.last_post_at, t.created_at
		FROM forum_topics t
		LEFT JOIN users u ON u.id = t.author_id
		WHERE t.section_id = $1
		ORDER BY t.last_post_at DESC NULLS LAST, t.id DESC
	`, sectionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []TopicView
	for rows.Next() {
		var t TopicView
		if err := rows.Scan(&t.ID, &t.Title, &t.AuthorID, &t.AuthorUsername,
			&t.PostCount, &t.LastPostAt, &t.CreatedAt); err != nil {
			return nil, err
		}
		result = append(result, t)
	}
	if result == nil {
		result = []TopicView{}
	}
	return result, rows.Err()
}

// GetTopic возвращает тему по ID, или (nil, nil) если не найдена.
func (r *Repo) GetTopic(ctx context.Context, id int64) (*Topic, error) {
	var t Topic
	err := r.pool.QueryRow(ctx, `
		SELECT id, section_id, title, author_id, post_count, last_post_at, last_post_by, created_at
		FROM forum_topics
		WHERE id = $1
	`, id).Scan(&t.ID, &t.SectionID, &t.Title, &t.AuthorID, &t.PostCount, &t.LastPostAt, &t.LastPostBy, &t.CreatedAt)

	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &t, nil
}

// ListPostsByTopic возвращает посты темы, отсортированные по created_at ASC, лимит 500.
func (r *Repo) ListPostsByTopic(ctx context.Context, topicID int64) ([]PostView, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT p.id, p.topic_id, p.author_id, COALESCE(u.username, '[удалён]') AS author_username,
		       p.parent_id, p.parent_author_snapshot, p.parent_content_snapshot, p.content, p.created_at
		FROM forum_posts p
		LEFT JOIN users u ON u.id = p.author_id
		WHERE p.topic_id = $1
		ORDER BY p.created_at ASC, p.id ASC
		LIMIT 500
	`, topicID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []PostView
	for rows.Next() {
		var p PostView
		if err := rows.Scan(&p.ID, &p.TopicID, &p.AuthorID, &p.AuthorUsername,
			&p.ParentID, &p.ParentAuthorSnapshot, &p.ParentContentSnapshot, &p.Content, &p.CreatedAt); err != nil {
			return nil, err
		}
		result = append(result, p)
	}
	if result == nil {
		result = []PostView{}
	}
	return result, rows.Err()
}

// CreateSection создаёт раздел и возвращает его ID.
func (r *Repo) CreateSection(ctx context.Context, s *Section) (int64, error) {
	var id int64
	err := r.pool.QueryRow(ctx, `
		INSERT INTO forum_sections (title, description, sort_order)
		VALUES ($1, $2, $3)
		RETURNING id
	`, s.Title, s.Description, s.SortOrder).Scan(&id)

	if err != nil {
		return 0, err
	}
	return id, nil
}

// CreateTopic создаёт тему и возвращает её ID. Маппит PG-ошибку 23503 (FK) в ErrSectionNotFound.
func (r *Repo) CreateTopic(ctx context.Context, t *Topic) (int64, error) {
	var id int64
	err := r.pool.QueryRow(ctx, `
		INSERT INTO forum_topics (section_id, title, author_id)
		VALUES ($1, $2, $3)
		RETURNING id
	`, t.SectionID, t.Title, t.AuthorID).Scan(&id)

	if err != nil {
		if pgErr := (*pgconn.PgError)(nil); errors.As(err, &pgErr) {
			if pgErr.Code == "23503" { // FK violation
				return 0, ErrSectionNotFound
			}
		}
		return 0, err
	}
	return id, nil
}

// CreatePost создаёт пост. Если ParentID != nil, открывает транзакцию для проверки родителя и snapshot.
func (r *Repo) CreatePost(ctx context.Context, p *Post) (int64, error) {
	if p.ParentID == nil {
		// Простой случай: корневой пост
		var id int64
		err := r.pool.QueryRow(ctx, `
			INSERT INTO forum_posts (topic_id, author_id, content)
			VALUES ($1, $2, $3)
			RETURNING id
		`, p.TopicID, p.AuthorID, p.Content).Scan(&id)

		if err != nil {
			if pgErr := (*pgconn.PgError)(nil); errors.As(err, &pgErr) {
				if pgErr.Code == "23503" { // FK violation
					return 0, ErrTopicNotFound
				}
			}
			return 0, err
		}
		return id, nil
	}

	// Сложный случай: ответ с цитатой
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback(ctx)

	// Проверяем существование и корректность родителя
	var parentTopicID int64
	var parentParentID *int64
	var parentAuthorID int64
	var parentContent string

	err = tx.QueryRow(ctx, `
		SELECT topic_id, parent_id, author_id, content
		FROM forum_posts
		WHERE id = $1
	`, p.ParentID).Scan(&parentTopicID, &parentParentID, &parentAuthorID, &parentContent)

	if err == pgx.ErrNoRows {
		return 0, ErrParentNotFound
	}
	if err != nil {
		return 0, err
	}

	// Проверяем, что родитель из той же темы
	if parentTopicID != p.TopicID {
		return 0, ErrParentNotFound
	}

	// Проверяем, что это не ответ на ответ (depth ≤ 1)
	if parentParentID != nil {
		return 0, ErrReplyToReply
	}

	// Заполняем snapshot: первые 100 рун текста родителя
	contentRunes := []rune(parentContent)
	if len(contentRunes) > 100 {
		contentRunes = contentRunes[:100]
	}

	// Получаем username родителя
	var parentAuthorUsername string
	err = tx.QueryRow(ctx, `SELECT username FROM users WHERE id = $1`, parentAuthorID).
		Scan(&parentAuthorUsername)
	if err != nil {
		if err == pgx.ErrNoRows {
			parentAuthorUsername = "[удалён]"
		} else {
			return 0, err
		}
	}

	p.ParentAuthorSnapshot = &parentAuthorUsername
	contentSnapshot := string(contentRunes)
	p.ParentContentSnapshot = &contentSnapshot

	// Вставляем пост
	var id int64
	err = tx.QueryRow(ctx, `
		INSERT INTO forum_posts (topic_id, author_id, parent_id, parent_author_snapshot, parent_content_snapshot, content)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id
	`, p.TopicID, p.AuthorID, p.ParentID, p.ParentAuthorSnapshot, p.ParentContentSnapshot, p.Content).Scan(&id)

	if err != nil {
		if pgErr := (*pgconn.PgError)(nil); errors.As(err, &pgErr) {
			if pgErr.Code == "23503" { // FK violation
				// Проверяем, какая FK нарушена (topic_id или parent_id)
				if pgErr.ConstraintName == "forum_posts_topic_id_fkey" {
					return 0, ErrTopicNotFound
				}
				return 0, ErrParentNotFound
			}
		}
		return 0, err
	}

	if err := tx.Commit(ctx); err != nil {
		return 0, err
	}

	return id, nil
}

// ListPostsAfter возвращает до 50 постов темы с id > afterID, по возрастанию id.
// Используется для SSE catch-up (Last-Event-ID).
func (r *Repo) ListPostsAfter(ctx context.Context, topicID, afterID int64) ([]PostView, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT p.id, p.topic_id, p.author_id, COALESCE(u.username, '[удалён]') AS author_username,
		       p.parent_id, p.parent_author_snapshot, p.parent_content_snapshot, p.content, p.created_at
		FROM forum_posts p
		LEFT JOIN users u ON u.id = p.author_id
		WHERE p.topic_id = $1 AND p.id > $2
		ORDER BY p.id ASC
		LIMIT 50
	`, topicID, afterID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []PostView
	for rows.Next() {
		var p PostView
		if err := rows.Scan(&p.ID, &p.TopicID, &p.AuthorID, &p.AuthorUsername,
			&p.ParentID, &p.ParentAuthorSnapshot, &p.ParentContentSnapshot, &p.Content, &p.CreatedAt); err != nil {
			return nil, err
		}
		result = append(result, p)
	}
	if result == nil {
		result = []PostView{}
	}
	return result, rows.Err()
}

// UpdateSection обновляет название и описание раздела.
func (r *Repo) UpdateSection(ctx context.Context, id int64, title, description string) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE forum_sections SET title = $1, description = $2 WHERE id = $3
	`, title, description, id)
	return err
}

// UpdateTopic обновляет название темы.
func (r *Repo) UpdateTopic(ctx context.Context, id int64, title string) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE forum_topics SET title = $1 WHERE id = $2
	`, title, id)
	return err
}

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

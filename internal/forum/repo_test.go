//go:build integration

package forum_test

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/Pegorino82/lfcru_forum/internal/forum"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pressly/goose/v3"

	_ "github.com/jackc/pgx/v5/stdlib"
)

const migrationsPath = "../../migrations"

func setupPool(t *testing.T) *pgxpool.Pool {
	t.Helper()
	url := os.Getenv("DATABASE_URL")
	if url == "" {
		t.Fatal("DATABASE_URL not set")
	}
	pool, err := pgxpool.New(context.Background(), url)
	if err != nil {
		t.Fatalf("failed to connect: %v", err)
	}
	t.Cleanup(pool.Close)

	db, err := goose.OpenDBWithDriver("pgx", url)
	if err != nil {
		t.Fatalf("goose open: %v", err)
	}
	defer db.Close()
	if err := goose.SetDialect("postgres"); err != nil {
		t.Fatalf("goose dialect: %v", err)
	}
	if err := goose.Up(db, migrationsPath); err != nil && !strings.Contains(err.Error(), "no migration files found") {
		t.Fatalf("goose up: %v", err)
	}

	return pool
}

func insertUser(t *testing.T, pool *pgxpool.Pool, email, username string) int64 {
	t.Helper()
	var id int64
	err := pool.QueryRow(context.Background(),
		`INSERT INTO users (email, username, pass_hash) VALUES ($1, $2, '\x68617368') RETURNING id`,
		email, username,
	).Scan(&id)
	if err != nil {
		err2 := pool.QueryRow(context.Background(),
			`SELECT id FROM users WHERE email = $1`, email,
		).Scan(&id)
		if err2 != nil {
			t.Fatalf("insertUser: %v / %v", err, err2)
		}
	}
	return id
}

func insertSection(t *testing.T, pool *pgxpool.Pool) int64 {
	t.Helper()
	var id int64
	err := pool.QueryRow(context.Background(),
		`INSERT INTO forum_sections (title) VALUES ('Test Section') RETURNING id`,
	).Scan(&id)
	if err != nil {
		t.Fatalf("insertSection: %v", err)
	}
	return id
}

func insertTopic(t *testing.T, pool *pgxpool.Pool, sectionID, authorID int64, title string) int64 {
	t.Helper()
	var id int64
	err := pool.QueryRow(context.Background(),
		`INSERT INTO forum_topics (section_id, author_id, title) VALUES ($1, $2, $3) RETURNING id`,
		sectionID, authorID, title,
	).Scan(&id)
	if err != nil {
		t.Fatalf("insertTopic: %v", err)
	}
	return id
}

func insertPost(t *testing.T, pool *pgxpool.Pool, topicID, authorID int64, content string) {
	t.Helper()
	_, err := pool.Exec(context.Background(),
		`INSERT INTO forum_posts (topic_id, author_id, content) VALUES ($1, $2, $3)`,
		topicID, authorID, content,
	)
	if err != nil {
		t.Fatalf("insertPost: %v", err)
	}
}

func cleanForum(t *testing.T, pool *pgxpool.Pool) {
	t.Helper()
	ctx := context.Background()
	pool.Exec(ctx, `DELETE FROM forum_posts`)
	pool.Exec(ctx, `DELETE FROM forum_topics`)
	pool.Exec(ctx, `DELETE FROM forum_sections`)
}

func TestLatestActive_Empty(t *testing.T) {
	pool := setupPool(t)
	cleanForum(t, pool)
	defer cleanForum(t, pool)

	repo := forum.NewRepo(pool)
	result, err := repo.LatestActive(context.Background(), 5)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 0 {
		t.Errorf("expected empty slice, got %d", len(result))
	}
}

func TestLatestActive_Limit(t *testing.T) {
	pool := setupPool(t)
	cleanForum(t, pool)
	defer cleanForum(t, pool)

	ctx := context.Background()
	authorID := insertUser(t, pool, "forumtest@example.com", "forumtest")
	sectionID := insertSection(t, pool)

	for i := 0; i < 7; i++ {
		topicID := insertTopic(t, pool, sectionID, authorID, "topic")
		insertPost(t, pool, topicID, authorID, "content")
	}

	repo := forum.NewRepo(pool)
	result, err := repo.LatestActive(ctx, 5)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 5 {
		t.Errorf("expected 5, got %d", len(result))
	}
}

func TestLatestActive_ExcludesTopicsWithoutPosts(t *testing.T) {
	pool := setupPool(t)
	cleanForum(t, pool)
	defer cleanForum(t, pool)

	ctx := context.Background()
	authorID := insertUser(t, pool, "forumtest2@example.com", "forumtest2")
	sectionID := insertSection(t, pool)

	// Тема с постом
	topicWithPost := insertTopic(t, pool, sectionID, authorID, "active topic")
	insertPost(t, pool, topicWithPost, authorID, "content")

	// Тема без поста
	insertTopic(t, pool, sectionID, authorID, "empty topic")

	repo := forum.NewRepo(pool)
	result, err := repo.LatestActive(ctx, 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 1 {
		t.Errorf("expected 1, got %d", len(result))
	}
	if result[0].Title != "active topic" {
		t.Errorf("expected 'active topic', got %q", result[0].Title)
	}
}

func TestLatestActive_SortedDesc(t *testing.T) {
	pool := setupPool(t)
	cleanForum(t, pool)
	defer cleanForum(t, pool)

	ctx := context.Background()
	authorID := insertUser(t, pool, "forumtest3@example.com", "forumtest3")
	sectionID := insertSection(t, pool)

	topic1 := insertTopic(t, pool, sectionID, authorID, "first")
	topic2 := insertTopic(t, pool, sectionID, authorID, "second")

	insertPost(t, pool, topic1, authorID, "post1")
	time.Sleep(10 * time.Millisecond)
	insertPost(t, pool, topic2, authorID, "post2")

	repo := forum.NewRepo(pool)
	result, err := repo.LatestActive(ctx, 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) < 2 {
		t.Fatalf("expected at least 2 results")
	}
	if result[0].Title != "second" {
		t.Errorf("expected 'second' first (most recent), got %q", result[0].Title)
	}
}

func TestLatestActive_LastPostByName(t *testing.T) {
	pool := setupPool(t)
	cleanForum(t, pool)
	defer cleanForum(t, pool)

	ctx := context.Background()
	authorID := insertUser(t, pool, "forumtest4@example.com", "forumtest4")
	sectionID := insertSection(t, pool)
	topicID := insertTopic(t, pool, sectionID, authorID, "topic with name")
	insertPost(t, pool, topicID, authorID, "content")

	repo := forum.NewRepo(pool)
	result, err := repo.LatestActive(ctx, 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) == 0 {
		t.Fatal("expected at least 1 result")
	}
	if result[0].LastPostByName != "forumtest4" {
		t.Errorf("expected 'forumtest4', got %q", result[0].LastPostByName)
	}
}

func TestLatestActive_DeletedUserShowsPlaceholder(t *testing.T) {
	pool := setupPool(t)
	cleanForum(t, pool)
	defer cleanForum(t, pool)

	ctx := context.Background()
	authorID := insertUser(t, pool, "tobedeleted@example.com", "tobedeleted")
	sectionID := insertSection(t, pool)
	topicID := insertTopic(t, pool, sectionID, authorID, "topic orphan")
	insertPost(t, pool, topicID, authorID, "content")

	// Обновим last_post_by на несуществующий id (эмулируем SET NULL)
	_, err := pool.Exec(ctx, `UPDATE forum_topics SET last_post_by = NULL WHERE id = $1`, topicID)
	if err != nil {
		t.Fatalf("update last_post_by: %v", err)
	}

	repo := forum.NewRepo(pool)
	result, err := repo.LatestActive(ctx, 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	found := false
	for _, r := range result {
		if r.ID == topicID {
			found = true
			if r.LastPostByName != "[удалён]" {
				t.Errorf("expected '[удалён]', got %q", r.LastPostByName)
			}
		}
	}
	if !found {
		t.Error("topic not found in results")
	}
}

func TestLatestActive_TriggerUpdatesFields(t *testing.T) {
	pool := setupPool(t)
	cleanForum(t, pool)
	defer cleanForum(t, pool)

	ctx := context.Background()
	authorID := insertUser(t, pool, "triggertest@example.com", "triggertest")
	sectionID := insertSection(t, pool)
	topicID := insertTopic(t, pool, sectionID, authorID, "trigger topic")

	// До поста — last_post_at IS NULL, post_count = 0
	var postCount int
	pool.QueryRow(ctx, `SELECT post_count FROM forum_topics WHERE id = $1`, topicID).Scan(&postCount)
	if postCount != 0 {
		t.Errorf("expected post_count=0 before insert, got %d", postCount)
	}

	insertPost(t, pool, topicID, authorID, "first post")

	// После поста — триггер должен обновить поля
	var lastPostBy *int64
	pool.QueryRow(ctx, `SELECT post_count, last_post_by FROM forum_topics WHERE id = $1`, topicID).Scan(&postCount, &lastPostBy)
	if postCount != 1 {
		t.Errorf("expected post_count=1 after insert, got %d", postCount)
	}
	if lastPostBy == nil || *lastPostBy != authorID {
		t.Errorf("expected last_post_by=%d, got %v", authorID, lastPostBy)
	}
}

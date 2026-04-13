//go:build integration

package news_test

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/Pegorino82/lfcru_forum/internal/news"
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

func insertUser(t *testing.T, pool *pgxpool.Pool) int64 {
	t.Helper()
	var id int64
	err := pool.QueryRow(context.Background(),
		`INSERT INTO users (email, username, pass_hash) VALUES ($1, $2, $3) RETURNING id`,
		"newstest@example.com", "newstest", []byte("hash"),
	).Scan(&id)
	if err != nil {
		// Пользователь уже существует — получим его id
		err2 := pool.QueryRow(context.Background(),
			`SELECT id FROM users WHERE email = $1`, "newstest@example.com",
		).Scan(&id)
		if err2 != nil {
			t.Fatalf("failed to get user: %v", err2)
		}
	}
	return id
}

func cleanNews(t *testing.T, pool *pgxpool.Pool) {
	t.Helper()
	_, err := pool.Exec(context.Background(), `DELETE FROM news WHERE author_id IN (SELECT id FROM users WHERE email = 'newstest@example.com')`)
	if err != nil {
		t.Fatalf("cleanNews: %v", err)
	}
}

func TestLatestPublished_Empty(t *testing.T) {
	pool := setupPool(t)
	authorID := insertUser(t, pool)
	cleanNews(t, pool)
	defer cleanNews(t, pool)

	repo := news.NewRepo(pool)
	result, err := repo.LatestPublished(context.Background(), 5)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 0 {
		t.Errorf("expected empty slice, got %d items", len(result))
	}
	_ = authorID
}

func TestLatestPublished_Limit(t *testing.T) {
	pool := setupPool(t)
	authorID := insertUser(t, pool)
	cleanNews(t, pool)
	defer cleanNews(t, pool)

	ctx := context.Background()
	now := time.Now()
	for i := 0; i < 7; i++ {
		pub := now.Add(time.Duration(-i) * time.Hour)
		_, err := pool.Exec(ctx,
			`INSERT INTO news (title, is_published, author_id, published_at) VALUES ($1, true, $2, $3)`,
			"news title", authorID, pub,
		)
		if err != nil {
			t.Fatalf("insert news: %v", err)
		}
	}

	repo := news.NewRepo(pool)
	result, err := repo.LatestPublished(ctx, 5)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 5 {
		t.Errorf("expected 5 items, got %d", len(result))
	}
}

func TestLatestPublished_ExcludesDrafts(t *testing.T) {
	pool := setupPool(t)
	authorID := insertUser(t, pool)
	cleanNews(t, pool)
	defer cleanNews(t, pool)

	ctx := context.Background()
	now := time.Now()
	// Опубликованная
	_, err := pool.Exec(ctx,
		`INSERT INTO news (title, is_published, author_id, published_at) VALUES ($1, true, $2, $3)`,
		"published", authorID, now,
	)
	if err != nil {
		t.Fatalf("insert published: %v", err)
	}
	// Черновик
	_, err = pool.Exec(ctx,
		`INSERT INTO news (title, is_published, author_id) VALUES ($1, false, $2)`,
		"draft", authorID,
	)
	if err != nil {
		t.Fatalf("insert draft: %v", err)
	}

	repo := news.NewRepo(pool)
	result, err := repo.LatestPublished(ctx, 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for _, n := range result {
		if n.Title == "draft" {
			t.Error("draft should not be returned")
		}
	}
	if len(result) != 1 {
		t.Errorf("expected 1 item, got %d", len(result))
	}
}

func TestGetPublishedByID_Found(t *testing.T) {
	pool := setupPool(t)
	authorID := insertUser(t, pool)
	cleanNews(t, pool)
	defer cleanNews(t, pool)

	ctx := context.Background()
	var newsID int64
	err := pool.QueryRow(ctx,
		`INSERT INTO news (title, content, is_published, author_id, published_at) VALUES ($1, $2, true, $3, now()) RETURNING id`,
		"Тестовая статья", "Полный текст статьи", authorID,
	).Scan(&newsID)
	if err != nil {
		t.Fatalf("insert news: %v", err)
	}

	repo := news.NewRepo(pool)
	n, err := repo.GetPublishedByID(ctx, newsID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if n == nil {
		t.Fatal("expected *News, got nil")
	}
	if n.ID != newsID {
		t.Errorf("expected ID %d, got %d", newsID, n.ID)
	}
	if n.Title != "Тестовая статья" {
		t.Errorf("unexpected title: %q", n.Title)
	}
	if n.Content != "Полный текст статьи" {
		t.Errorf("unexpected content: %q", n.Content)
	}
}

func TestGetPublishedByID_NotFound(t *testing.T) {
	pool := setupPool(t)
	insertUser(t, pool)
	cleanNews(t, pool)
	defer cleanNews(t, pool)

	repo := news.NewRepo(pool)
	n, err := repo.GetPublishedByID(context.Background(), 999999)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if n != nil {
		t.Errorf("expected nil, got %+v", n)
	}
}

func TestGetPublishedByID_Draft(t *testing.T) {
	pool := setupPool(t)
	authorID := insertUser(t, pool)
	cleanNews(t, pool)
	defer cleanNews(t, pool)

	ctx := context.Background()
	var newsID int64
	err := pool.QueryRow(ctx,
		`INSERT INTO news (title, is_published, author_id) VALUES ($1, false, $2) RETURNING id`,
		"черновик", authorID,
	).Scan(&newsID)
	if err != nil {
		t.Fatalf("insert draft: %v", err)
	}

	repo := news.NewRepo(pool)
	n, err := repo.GetPublishedByID(ctx, newsID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if n != nil {
		t.Errorf("draft should return nil, got %+v", n)
	}
}

func TestListPublished_Empty(t *testing.T) {
	pool := setupPool(t)
	insertUser(t, pool)
	cleanNews(t, pool)
	defer cleanNews(t, pool)

	repo := news.NewRepo(pool)
	items, total, err := repo.ListPublished(context.Background(), 20, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if total != 0 {
		t.Errorf("expected total=0, got %d", total)
	}
	if len(items) != 0 {
		t.Errorf("expected empty slice, got %d items", len(items))
	}
}

func TestListPublished_Pagination(t *testing.T) {
	pool := setupPool(t)
	authorID := insertUser(t, pool)
	cleanNews(t, pool)
	defer cleanNews(t, pool)

	ctx := context.Background()
	now := time.Now()
	for i := 0; i < 25; i++ {
		pub := now.Add(time.Duration(-i) * time.Hour)
		_, err := pool.Exec(ctx,
			`INSERT INTO news (title, is_published, author_id, published_at) VALUES ($1, true, $2, $3)`,
			"test-list-news", authorID, pub,
		)
		if err != nil {
			t.Fatalf("insert news: %v", err)
		}
	}

	repo := news.NewRepo(pool)

	// Page 1: first 20
	items, total, err := repo.ListPublished(ctx, 20, 0)
	if err != nil {
		t.Fatalf("page 1: %v", err)
	}
	if total != 25 {
		t.Errorf("expected total=25, got %d", total)
	}
	if len(items) != 20 {
		t.Errorf("expected 20 items on page 1, got %d", len(items))
	}

	// Page 2: remaining 5
	items2, _, err := repo.ListPublished(ctx, 20, 20)
	if err != nil {
		t.Fatalf("page 2: %v", err)
	}
	if len(items2) != 5 {
		t.Errorf("expected 5 items on page 2, got %d", len(items2))
	}
}

func TestListPublished_ExcludesDrafts(t *testing.T) {
	pool := setupPool(t)
	authorID := insertUser(t, pool)
	cleanNews(t, pool)
	defer cleanNews(t, pool)

	ctx := context.Background()
	_, _ = pool.Exec(ctx,
		`INSERT INTO news (title, is_published, author_id, published_at) VALUES ($1, true, $2, now())`,
		"test-list-published", authorID,
	)
	_, _ = pool.Exec(ctx,
		`INSERT INTO news (title, is_published, author_id) VALUES ($1, false, $2)`,
		"test-list-draft", authorID,
	)

	repo := news.NewRepo(pool)
	items, total, err := repo.ListPublished(ctx, 20, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if total != 1 {
		t.Errorf("expected total=1, got %d", total)
	}
	for _, item := range items {
		if item.Title == "test-list-draft" {
			t.Error("draft should not appear in list")
		}
	}
}

func TestListPublished_SortedDesc(t *testing.T) {
	pool := setupPool(t)
	authorID := insertUser(t, pool)
	cleanNews(t, pool)
	defer cleanNews(t, pool)

	ctx := context.Background()
	base := time.Now()
	for i, title := range []string{"test-list-oldest", "test-list-middle", "test-list-newest"} {
		pub := base.Add(time.Duration(i) * time.Hour)
		_, _ = pool.Exec(ctx,
			`INSERT INTO news (title, is_published, author_id, published_at) VALUES ($1, true, $2, $3)`,
			title, authorID, pub,
		)
	}

	repo := news.NewRepo(pool)
	items, _, err := repo.ListPublished(ctx, 20, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(items) < 3 {
		t.Fatalf("expected at least 3 items")
	}
	if items[0].Title != "test-list-newest" {
		t.Errorf("first item should be newest, got %q", items[0].Title)
	}
}

func TestLatestPublished_SortedDesc(t *testing.T) {
	pool := setupPool(t)
	authorID := insertUser(t, pool)
	cleanNews(t, pool)
	defer cleanNews(t, pool)

	ctx := context.Background()
	base := time.Now()
	titles := []string{"oldest", "middle", "newest"}
	for i, title := range titles {
		pub := base.Add(time.Duration(i) * time.Hour)
		_, err := pool.Exec(ctx,
			`INSERT INTO news (title, is_published, author_id, published_at) VALUES ($1, true, $2, $3)`,
			title, authorID, pub,
		)
		if err != nil {
			t.Fatalf("insert: %v", err)
		}
	}

	repo := news.NewRepo(pool)
	result, err := repo.LatestPublished(ctx, 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) < 3 {
		t.Fatalf("expected at least 3 items")
	}
	if result[0].Title != "newest" {
		t.Errorf("first item should be newest, got %q", result[0].Title)
	}
}

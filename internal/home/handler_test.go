//go:build integration

package home_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/Pegorino82/lfcru_forum/internal/auth"
	"github.com/Pegorino82/lfcru_forum/internal/forum"
	"github.com/Pegorino82/lfcru_forum/internal/home"
	appMiddleware "github.com/Pegorino82/lfcru_forum/internal/middleware"
	"github.com/Pegorino82/lfcru_forum/internal/match"
	"github.com/Pegorino82/lfcru_forum/internal/news"
	"github.com/Pegorino82/lfcru_forum/internal/ratelimit"
	"github.com/Pegorino82/lfcru_forum/internal/session"
	"github.com/Pegorino82/lfcru_forum/internal/tmpl"
	"github.com/Pegorino82/lfcru_forum/internal/user"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/pressly/goose/v3"
	"golang.org/x/crypto/bcrypt"

	_ "github.com/jackc/pgx/v5/stdlib"
)

const migrationsPath = "../../migrations"
const templatesPath = "../../templates"

var (
	dbOnce     sync.Once
	sharedPool *pgxpool.Pool
	dbSetupErr error
)

func testDB(t *testing.T) *pgxpool.Pool {
	t.Helper()
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		t.Skip("DATABASE_URL not set, skipping integration test")
	}
	dbOnce.Do(func() {
		pool, err := pgxpool.New(context.Background(), dbURL)
		if err != nil {
			dbSetupErr = err
			return
		}
		if err := pool.Ping(context.Background()); err != nil {
			pool.Close()
			dbSetupErr = err
			return
		}
		db, err := goose.OpenDBWithDriver("pgx", dbURL)
		if err != nil {
			pool.Close()
			dbSetupErr = err
			return
		}
		defer db.Close()
		if err := goose.SetDialect("postgres"); err != nil {
			pool.Close()
			dbSetupErr = err
			return
		}
		if err := goose.Up(db, migrationsPath); err != nil {
			pool.Close()
			dbSetupErr = err
			return
		}
		sharedPool = pool
	})
	if dbSetupErr != nil {
		t.Fatalf("db setup: %v", dbSetupErr)
	}
	return sharedPool
}

func cleanAll(t *testing.T, pool *pgxpool.Pool) {
	t.Helper()
	ctx := context.Background()
	pool.Exec(ctx, `DELETE FROM forum_posts`)
	pool.Exec(ctx, `DELETE FROM forum_topics`)
	pool.Exec(ctx, `DELETE FROM forum_sections`)
	pool.Exec(ctx, `DELETE FROM matches`)
	pool.Exec(ctx, `DELETE FROM news`)
}

func newTestServer(t *testing.T, pool *pgxpool.Pool) *echo.Echo {
	t.Helper()
	renderer, err := tmpl.New(os.DirFS(templatesPath), "templates/")
	if err != nil {
		t.Fatalf("load templates: %v", err)
	}

	svc := auth.NewService(
		user.NewRepo(pool),
		session.NewRepo(pool),
		ratelimit.NewLoginAttemptRepo(pool),
		auth.Config{
			BcryptCost:         bcrypt.MinCost,
			SessionLifetime:    30 * 24 * time.Hour,
			RateLimitWindow:    10 * time.Minute,
			RateLimitMax:       5,
			SessionGracePeriod: 5 * time.Minute,
			MaxSessionsPerUser: 10,
		},
	)

	e := echo.New()
	e.HideBanner = true
	e.Renderer = renderer
	e.Use(middleware.Recover())
	e.Use(appMiddleware.CSRFMiddleware())
	e.Use(auth.LoadSession(svc))

	homeHandler := home.NewHandler(news.NewRepo(pool), match.NewRepo(pool), forum.NewRepo(pool))
	e.GET("/", homeHandler.ShowHome)

	return e
}

func doGet(t *testing.T, e *echo.Echo, path string, opts ...func(*http.Request)) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(http.MethodGet, path, nil)
	for _, opt := range opts {
		opt(req)
	}
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	return rec
}

func withHTMX(req *http.Request) {
	req.Header.Set("HX-Request", "true")
}

// ─── Tests ────────────────────────────────────────────────────────────────────

func TestHomeHandler_EmptyState(t *testing.T) {
	pool := testDB(t)
	cleanAll(t, pool)
	defer cleanAll(t, pool)

	e := newTestServer(t, pool)
	rec := doGet(t, e, "/")

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	body := rec.Body.String()

	emptyStates := []string{
		"На сайте еще не добавлены новости",
		"Ближайших матчей нет",
		"В форуме пока нет активных обсуждений",
	}
	for _, msg := range emptyStates {
		if !strings.Contains(body, msg) {
			t.Errorf("expected empty-state %q in body", msg)
		}
	}
}

func TestHomeHandler_WithData(t *testing.T) {
	pool := testDB(t)
	cleanAll(t, pool)
	defer cleanAll(t, pool)

	ctx := context.Background()

	// Создаём пользователя для новостей и форума
	var authorID int64
	err := pool.QueryRow(ctx,
		`INSERT INTO users (email, username, pass_hash) VALUES ('hometest@example.com', 'hometest', '\x68617368') RETURNING id`,
	).Scan(&authorID)
	if err != nil {
		pool.QueryRow(ctx, `SELECT id FROM users WHERE email='hometest@example.com'`).Scan(&authorID)
	}

	// Новость
	now := time.Now()
	_, err = pool.Exec(ctx,
		`INSERT INTO news (title, status, author_id, published_at) VALUES ($1, 'published', $2, $3)`,
		"Ливерпуль победил!", authorID, now,
	)
	if err != nil {
		t.Fatalf("insert news: %v", err)
	}

	// Матч
	_, err = pool.Exec(ctx,
		`INSERT INTO matches (opponent, match_date, tournament) VALUES ($1, $2, $3)`,
		"Манчестер Юнайтед", now.Add(48*time.Hour), "АПЛ",
	)
	if err != nil {
		t.Fatalf("insert match: %v", err)
	}

	// Форум
	var sectionID, topicID int64
	pool.QueryRow(ctx, `INSERT INTO forum_sections (title) VALUES ('Общий') RETURNING id`).Scan(&sectionID)
	pool.QueryRow(ctx,
		`INSERT INTO forum_topics (section_id, author_id, title) VALUES ($1, $2, $3) RETURNING id`,
		sectionID, authorID, "Обсуждение матча",
	).Scan(&topicID)
	pool.Exec(ctx,
		`INSERT INTO forum_posts (topic_id, author_id, content) VALUES ($1, $2, $3)`,
		topicID, authorID, "Отличная игра!",
	)

	e := newTestServer(t, pool)
	rec := doGet(t, e, "/")

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	body := rec.Body.String()

	if !strings.Contains(body, "Ливерпуль победил!") {
		t.Errorf("expected news title in body")
	}
	if !strings.Contains(body, "Манчестер Юнайтед") {
		t.Errorf("expected match opponent in body")
	}
	if !strings.Contains(body, "Обсуждение матча") {
		t.Errorf("expected forum topic in body")
	}
}

func TestHomeHandler_HTMXPartial(t *testing.T) {
	pool := testDB(t)
	cleanAll(t, pool)
	defer cleanAll(t, pool)

	e := newTestServer(t, pool)
	rec := doGet(t, e, "/", withHTMX)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	body := rec.Body.String()

	if strings.Contains(body, "<html") {
		t.Error("HTMX partial should not contain <html>")
	}
	if strings.Contains(body, "<head>") {
		t.Error("HTMX partial should not contain <head>")
	}
}

func TestHomeHandler_GuestAccess(t *testing.T) {
	pool := testDB(t)
	cleanAll(t, pool)
	defer cleanAll(t, pool)

	e := newTestServer(t, pool)
	rec := doGet(t, e, "/") // без cookie

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 for guest, got %d", rec.Code)
	}
}

//go:build integration

package layout_test

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

// ─── DB setup ────────────────────────────────────────────────────────────────

var (
	dbOnce     sync.Once
	sharedPool *pgxpool.Pool
	dbSetupErr error
)

const migrationsPath = "../../migrations"
const templatesPath = "../../templates"

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

func truncateTables(t *testing.T, pool *pgxpool.Pool) {
	t.Helper()
	_, err := pool.Exec(context.Background(),
		"TRUNCATE login_attempts, sessions, users CASCADE")
	if err != nil {
		t.Fatalf("truncate: %v", err)
	}
}

func newSvc(pool *pgxpool.Pool) *auth.Service {
	return auth.NewService(
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
}

// ─── Echo server setup ───────────────────────────────────────────────────────

func newServer(t *testing.T, pool *pgxpool.Pool, svc *auth.Service) *echo.Echo {
	t.Helper()
	renderer, err := tmpl.New(os.DirFS(templatesPath), "templates/")
	if err != nil {
		t.Fatalf("load templates: %v", err)
	}

	e := echo.New()
	e.HideBanner = true
	e.Renderer = renderer
	e.Use(middleware.Recover())
	e.Use(appMiddleware.CSRFMiddleware())
	e.Use(auth.LoadSession(svc))

	homeHandler := home.NewHandler(news.NewRepo(pool), match.NewRepo(pool), forum.NewRepo(pool), nil)
	e.GET("/", homeHandler.ShowHome)
	auth.NewHandler(svc).RegisterRoutes(e)

	return e
}

// doRequest performs a GET request to the given path, optionally setting cookie and headers.
func doRequest(t *testing.T, e *echo.Echo, path string, opts ...func(*http.Request)) *httptest.ResponseRecorder {
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

func withSession(sessionID string) func(*http.Request) {
	return func(req *http.Request) {
		req.AddCookie(&http.Cookie{Name: "session_id", Value: sessionID})
	}
}

// ─── Tests ───────────────────────────────────────────────────────────────────

func TestLayoutStructure(t *testing.T) {
	pool := testDB(t)
	svc := newSvc(pool)
	truncateTables(t, pool)
	e := newServer(t, pool, svc)

	t.Run("header присутствует", func(t *testing.T) {
		rec := doRequest(t, e, "/login")
		body := rec.Body.String()
		if !strings.Contains(body, "<header>") {
			t.Errorf("expected <header> in response, got:\n%s", body)
		}
	})

	t.Run("nav с aria-label", func(t *testing.T) {
		rec := doRequest(t, e, "/login")
		body := rec.Body.String()
		if !strings.Contains(body, `aria-label="Основная навигация"`) {
			t.Errorf("expected aria-label in nav, got:\n%s", body)
		}
	})

	t.Run("footer присутствует", func(t *testing.T) {
		rec := doRequest(t, e, "/login")
		body := rec.Body.String()
		if !strings.Contains(body, "<footer>") {
			t.Errorf("expected <footer> in response, got:\n%s", body)
		}
	})

	t.Run("копирайт в footer", func(t *testing.T) {
		rec := doRequest(t, e, "/login")
		body := rec.Body.String()
		if !strings.Contains(body, "© 2026 LFC.ru") {
			t.Errorf("expected copyright in footer, got:\n%s", body)
		}
	})

	t.Run("дисклеймер в footer", func(t *testing.T) {
		rec := doRequest(t, e, "/login")
		body := rec.Body.String()
		if !strings.Contains(body, "Не является официальным сайтом Liverpool FC") {
			t.Errorf("expected disclaimer in footer, got:\n%s", body)
		}
	})

	t.Run("гостевой nav: ссылки на login и register", func(t *testing.T) {
		rec := doRequest(t, e, "/login")
		body := rec.Body.String()
		if !strings.Contains(body, "/login") || !strings.Contains(body, "/register") {
			t.Errorf("expected /login and /register links in nav, got:\n%s", body)
		}
	})

	t.Run("авторизованный nav: кнопка выхода", func(t *testing.T) {
		_, sess, err := svc.Register(context.Background(), auth.RegisterInput{
			Username:        "layoutuser",
			Email:           "layout@example.com",
			Password:        "password123",
			PasswordConfirm: "password123",
			IPAddr:          "127.0.0.1",
			UserAgent:       "go-test",
		})
		if err != nil {
			t.Fatalf("register: %v", err)
		}
		rec := doRequest(t, e, "/", withSession(sess.ID.String()))
		body := rec.Body.String()
		if !strings.Contains(body, `action="/logout"`) {
			t.Errorf("expected logout form in nav, got:\n%s", body)
		}
	})

	t.Run("HTMX partial login: нет header и footer", func(t *testing.T) {
		rec := doRequest(t, e, "/login", withHTMX)
		body := rec.Body.String()
		if strings.Contains(body, "<header>") {
			t.Errorf("expected no <header> in HTMX partial, got:\n%s", body)
		}
		if strings.Contains(body, "<footer>") {
			t.Errorf("expected no <footer> in HTMX partial, got:\n%s", body)
		}
	})

	t.Run("HTMX partial register: нет header и footer", func(t *testing.T) {
		rec := doRequest(t, e, "/register", withHTMX)
		body := rec.Body.String()
		if strings.Contains(body, "<header>") {
			t.Errorf("expected no <header> in HTMX partial, got:\n%s", body)
		}
		if strings.Contains(body, "<footer>") {
			t.Errorf("expected no <footer> in HTMX partial, got:\n%s", body)
		}
	})

	t.Run("skip-link target: id=content", func(t *testing.T) {
		rec := doRequest(t, e, "/login")
		body := rec.Body.String()
		if !strings.Contains(body, `id="content"`) {
			t.Errorf("expected id=\"content\" in response, got:\n%s", body)
		}
	})

	t.Run("skip-link: ссылка на #content", func(t *testing.T) {
		rec := doRequest(t, e, "/login")
		body := rec.Body.String()
		if !strings.Contains(body, `href="#content"`) {
			t.Errorf("expected skip-link href=\"#content\", got:\n%s", body)
		}
	})

	t.Run("семантический main", func(t *testing.T) {
		rec := doRequest(t, e, "/login")
		body := rec.Body.String()
		if !strings.Contains(body, "<main") {
			t.Errorf("expected <main in response, got:\n%s", body)
		}
	})
}

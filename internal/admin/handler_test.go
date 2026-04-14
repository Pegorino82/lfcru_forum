//go:build integration

package admin_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/Pegorino82/lfcru_forum/internal/admin"
	"github.com/Pegorino82/lfcru_forum/internal/auth"
	appMiddleware "github.com/Pegorino82/lfcru_forum/internal/middleware"
	"github.com/Pegorino82/lfcru_forum/internal/ratelimit"
	"github.com/Pegorino82/lfcru_forum/internal/session"
	"github.com/Pegorino82/lfcru_forum/internal/tmpl"
	"github.com/Pegorino82/lfcru_forum/internal/user"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/pressly/goose/v3"
	"golang.org/x/crypto/bcrypt"

	_ "github.com/jackc/pgx/v5/stdlib"
)

const (
	migrationsPath = "../../migrations"
	templatesPath  = "../../templates"
)

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

func cleanAdminData(t *testing.T, pool *pgxpool.Pool) {
	t.Helper()
	ctx := context.Background()
	pool.Exec(ctx, `DELETE FROM sessions WHERE user_id IN (SELECT id FROM users WHERE email LIKE 'admintest%')`)
	pool.Exec(ctx, `DELETE FROM users WHERE email LIKE 'admintest%'`)
}

// newTestServer builds an Echo server with admin routes registered.
func newTestServer(t *testing.T, pool *pgxpool.Pool) *echo.Echo {
	t.Helper()
	renderer, err := tmpl.New(os.DirFS(templatesPath), "templates/")
	if err != nil {
		t.Fatalf("load templates: %v", err)
	}

	userRepo := user.NewRepo(pool)
	authSvc := auth.NewService(
		userRepo,
		session.NewRepo(pool),
		ratelimit.NewLoginAttemptRepo(pool),
		auth.Config{
			BcryptCost:         bcrypt.MinCost,
			SessionLifetime:    30 * 24 * time.Hour,
			RateLimitWindow:    10 * time.Minute,
			RateLimitMax:       10,
			SessionGracePeriod: 5 * time.Minute,
			MaxSessionsPerUser: 10,
		},
	)

	e := echo.New()
	e.HideBanner = true
	e.Renderer = renderer
	e.Use(middleware.Recover())
	e.Use(appMiddleware.CSRFMiddleware())
	e.Use(auth.LoadSession(authSvc))

	adminGroup := e.Group("", admin.RequireAdminOrMod(renderer))
	adminGroup.GET("/admin", admin.NewHandler().Dashboard)

	return e
}

// createUserWithRole inserts a test user with the given role and returns (userID, sessionCookie).
func createUserWithRole(t *testing.T, pool *pgxpool.Pool, email, username, role string) (int64, string) {
	t.Helper()
	ctx := context.Background()
	hash, _ := bcrypt.GenerateFromPassword([]byte("password"), bcrypt.MinCost)
	var userID int64
	if err := pool.QueryRow(ctx,
		`INSERT INTO users (email, username, pass_hash, role) VALUES ($1, $2, $3, $4) RETURNING id`,
		email, username, hash, role,
	).Scan(&userID); err != nil {
		t.Fatalf("create user: %v", err)
	}
	sessID := uuid.New()
	if _, err := pool.Exec(ctx,
		`INSERT INTO sessions (id, user_id, ip_addr, user_agent, expires_at)
		 VALUES ($1, $2, '127.0.0.1', 'test-agent', now() + interval '30 days')`,
		sessID, userID,
	); err != nil {
		t.Fatalf("create session: %v", err)
	}
	return userID, sessID.String()
}

func doGet(t *testing.T, e *echo.Echo, path, sessID string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(http.MethodGet, path, nil)
	if sessID != "" {
		req.Header.Set("Cookie", "session_id="+sessID)
	}
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	return rec
}

// ─── Tests ───────────────────────────────────────────────────────────────────

// SC-01: Admin открывает /admin → 200 с дашбордом.
func TestDashboard_Admin(t *testing.T) {
	pool := testDB(t)
	cleanAdminData(t, pool)
	e := newTestServer(t, pool)

	_, sessID := createUserWithRole(t, pool, "admintest-admin@test.com", "admintest_admin", "admin")

	rec := doGet(t, e, "/admin", sessID)
	if rec.Code != http.StatusOK {
		t.Fatalf("want 200, got %d", rec.Code)
	}
	body := rec.Body.String()
	if body == "" {
		t.Fatal("expected non-empty response body")
	}
}

// SC-02: Moderator открывает /admin → 200 с дашбордом.
func TestDashboard_Moderator(t *testing.T) {
	pool := testDB(t)
	cleanAdminData(t, pool)
	e := newTestServer(t, pool)

	_, sessID := createUserWithRole(t, pool, "admintest-mod@test.com", "admintest_mod", "moderator")

	rec := doGet(t, e, "/admin", sessID)
	if rec.Code != http.StatusOK {
		t.Fatalf("want 200, got %d", rec.Code)
	}
}

// SC-03: Гость переходит на /admin → 302 на /login.
func TestDashboard_Guest(t *testing.T) {
	pool := testDB(t)
	e := newTestServer(t, pool)

	rec := doGet(t, e, "/admin", "")
	if rec.Code != http.StatusFound {
		t.Fatalf("want 302, got %d", rec.Code)
	}
	loc := rec.Header().Get("Location")
	if loc != "/login" {
		t.Fatalf("want redirect to /login, got %q", loc)
	}
}

// SC-04: User (без прав admin/mod) открывает /admin → 403.
func TestDashboard_RegularUser(t *testing.T) {
	pool := testDB(t)
	cleanAdminData(t, pool)
	e := newTestServer(t, pool)

	_, sessID := createUserWithRole(t, pool, "admintest-user@test.com", "admintest_user", "user")

	rec := doGet(t, e, "/admin", sessID)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("want 403, got %d", rec.Code)
	}
}

//go:build integration

package admin_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/Pegorino82/lfcru_forum/internal/admin"
	"github.com/Pegorino82/lfcru_forum/internal/auth"
	appMiddleware "github.com/Pegorino82/lfcru_forum/internal/middleware"
	"github.com/Pegorino82/lfcru_forum/internal/ratelimit"
	"github.com/Pegorino82/lfcru_forum/internal/session"
	"github.com/Pegorino82/lfcru_forum/internal/tmpl"
	"github.com/Pegorino82/lfcru_forum/internal/user"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"golang.org/x/crypto/bcrypt"
)

func newUsersAdminServer(t *testing.T, pool *pgxpool.Pool) *echo.Echo {
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

	userSvc := user.NewService(userRepo)
	usersAdminHandler := admin.NewUsersHandler(userSvc)

	e := echo.New()
	e.HideBanner = true
	e.Renderer = renderer
	e.Use(middleware.Recover())
	e.Use(appMiddleware.CSRFMiddleware())
	e.Use(auth.LoadSession(authSvc))

	adminGroup := e.Group("", admin.RequireAdminOrMod(renderer))
	adminGroup.GET("/admin/users", usersAdminHandler.List)
	adminGroup.POST("/admin/users/:id/ban", usersAdminHandler.Ban)
	adminGroup.POST("/admin/users/:id/unban", usersAdminHandler.Unban)

	return e
}

func cleanUsersAdminData(t *testing.T, pool *pgxpool.Pool) {
	t.Helper()
	ctx := context.Background()
	pool.Exec(ctx, `DELETE FROM sessions WHERE user_id IN (SELECT id FROM users WHERE email LIKE 'uadmintest%')`)
	pool.Exec(ctx, `DELETE FROM users WHERE email LIKE 'uadmintest%'`)
}

// getUserCsrf fetches the CSRF token by making a GET to the users list.
func getUsersCsrf(t *testing.T, e *echo.Echo, sessID string) string {
	t.Helper()
	rec := doGet(t, e, "/admin/users", sessID)
	for _, c := range rec.Result().Cookies() {
		if c.Name == "_csrf" {
			return c.Value
		}
	}
	t.Fatal("no _csrf cookie in response")
	return ""
}

// doUsersPost performs a POST with session + CSRF cookies and form body.
func doUsersPost(t *testing.T, e *echo.Echo, path string, sessID, csrfToken string) *httptest.ResponseRecorder {
	t.Helper()
	form := url.Values{"_csrf": {csrfToken}}
	req := httptest.NewRequest(http.MethodPost, path, strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Cookie", "session_id="+sessID+"; _csrf="+csrfToken)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	return rec
}

// ─── Tests ───────────────────────────────────────────────────────────────────

// SC-01: Admin открывает /admin/users → 200 со списком пользователей.
func TestAdminUsers_List(t *testing.T) {
	pool := testDB(t)
	cleanUsersAdminData(t, pool)
	e := newUsersAdminServer(t, pool)

	_, sessID := createUserWithRole(t, pool, "uadmintest-a1@test.com", "uadmintest_a1", "admin")

	rec := doGet(t, e, "/admin/users", sessID)
	if rec.Code != http.StatusOK {
		t.Fatalf("want 200, got %d\n%s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), "uadmintest_a1") {
		t.Fatal("response does not contain username")
	}
}

// EC-01: Admin банит пользователя → banned_at заполнена в БД.
func TestAdminUsers_BanUser(t *testing.T) {
	pool := testDB(t)
	cleanUsersAdminData(t, pool)
	e := newUsersAdminServer(t, pool)

	_, sessID := createUserWithRole(t, pool, "uadmintest-a2@test.com", "uadmintest_a2", "admin")
	targetID, _ := createUserWithRole(t, pool, "uadmintest-u2@test.com", "uadmintest_u2", "user")
	csrfToken := getUsersCsrf(t, e, sessID)

	rec := doUsersPost(t, e, "/admin/users/"+strconv.FormatInt(targetID, 10)+"/ban", sessID, csrfToken)
	if rec.Code != http.StatusSeeOther {
		t.Fatalf("want 303, got %d\n%s", rec.Code, rec.Body.String())
	}

	var bannedAt *string
	pool.QueryRow(context.Background(),
		`SELECT banned_at::text FROM users WHERE id = $1`, targetID,
	).Scan(&bannedAt)
	if bannedAt == nil {
		t.Fatal("banned_at should be set after ban")
	}
}

// EC-03: Admin разбанивает пользователя → banned_at = NULL.
func TestAdminUsers_UnbanUser(t *testing.T) {
	pool := testDB(t)
	cleanUsersAdminData(t, pool)
	e := newUsersAdminServer(t, pool)

	_, sessID := createUserWithRole(t, pool, "uadmintest-a3@test.com", "uadmintest_a3", "admin")
	targetID, _ := createUserWithRole(t, pool, "uadmintest-u3@test.com", "uadmintest_u3", "user")

	// Pre-ban the user directly in DB
	pool.Exec(context.Background(), `UPDATE users SET banned_at = now() WHERE id = $1`, targetID)

	csrfToken := getUsersCsrf(t, e, sessID)

	rec := doUsersPost(t, e, "/admin/users/"+strconv.FormatInt(targetID, 10)+"/unban", sessID, csrfToken)
	if rec.Code != http.StatusSeeOther {
		t.Fatalf("want 303, got %d\n%s", rec.Code, rec.Body.String())
	}

	var bannedAt *string
	pool.QueryRow(context.Background(),
		`SELECT banned_at::text FROM users WHERE id = $1`, targetID,
	).Scan(&bannedAt)
	if bannedAt != nil {
		t.Fatal("banned_at should be NULL after unban")
	}
}

// EC-04: Admin пытается забанить самого себя → 400.
func TestAdminUsers_BanSelf(t *testing.T) {
	pool := testDB(t)
	cleanUsersAdminData(t, pool)
	e := newUsersAdminServer(t, pool)

	adminID, sessID := createUserWithRole(t, pool, "uadmintest-a4@test.com", "uadmintest_a4", "admin")
	csrfToken := getUsersCsrf(t, e, sessID)

	rec := doUsersPost(t, e, "/admin/users/"+strconv.FormatInt(adminID, 10)+"/ban", sessID, csrfToken)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("want 400, got %d\n%s", rec.Code, rec.Body.String())
	}
}

// FM-02: Пользователь не найден → 404.
func TestAdminUsers_BanNotFound(t *testing.T) {
	pool := testDB(t)
	cleanUsersAdminData(t, pool)
	e := newUsersAdminServer(t, pool)

	_, sessID := createUserWithRole(t, pool, "uadmintest-a5@test.com", "uadmintest_a5", "admin")
	csrfToken := getUsersCsrf(t, e, sessID)

	rec := doUsersPost(t, e, "/admin/users/999999999/ban", sessID, csrfToken)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("want 404, got %d", rec.Code)
	}
}

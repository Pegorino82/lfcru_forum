//go:build integration

package auth_test

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"

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

const authTemplatesPath = "../../templates"

func newAuthTestServer(t *testing.T, pool *pgxpool.Pool) *echo.Echo {
	t.Helper()
	renderer, err := tmpl.New(os.DirFS(authTemplatesPath), "templates/")
	if err != nil {
		t.Fatalf("load templates: %v", err)
	}
	authSvc := auth.NewService(
		user.NewRepo(pool),
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
	auth.NewHandler(authSvc).RegisterRoutes(e)
	return e
}

func doHTMXPost(t *testing.T, e *echo.Echo, path string, form url.Values, csrfToken string) *httptest.ResponseRecorder {
	t.Helper()
	if csrfToken != "" {
		form.Set("_csrf", csrfToken)
	}
	req := httptest.NewRequest(http.MethodPost, path, strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("HX-Request", "true")
	if csrfToken != "" {
		req.Header.Set("Cookie", "_csrf="+csrfToken)
	}
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	return rec
}

func getCSRFCookie(t *testing.T, e *echo.Echo, path string) string {
	t.Helper()
	req := httptest.NewRequest(http.MethodGet, path, nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	for _, c := range rec.Result().Cookies() {
		if c.Name == "_csrf" {
			return c.Value
		}
	}
	t.Fatal("no _csrf cookie in GET response")
	return ""
}

// Regression test для FT-014: повторный HTMX POST /login с ошибкой не должен
// вкладывать форму в форму. Ответ partial должен содержать ровно один
// id="login-wrapper" и ровно один id="login-form".
func TestLogin_HTMX_InvalidCredentials_NoNestedForm(t *testing.T) {
	pool := testDB(t)
	truncate(t, pool)

	e := newAuthTestServer(t, pool)

	csrfToken := getCSRFCookie(t, e, "/login")

	form := url.Values{
		"email":    {"nobody@example.com"},
		"password": {"wrongpassword"},
	}

	// Первый неудачный логин
	rec1 := doHTMXPost(t, e, "/login", form, csrfToken)
	if rec1.Code != http.StatusUnprocessableEntity {
		t.Fatalf("first attempt: expected 422, got %d", rec1.Code)
	}
	body1 := rec1.Body.String()
	if strings.Count(body1, `id="login-wrapper"`) != 1 {
		t.Errorf("first attempt: expected exactly 1 id=\"login-wrapper\" in response, got %d\nbody:\n%s",
			strings.Count(body1, `id="login-wrapper"`), body1)
	}
	if strings.Count(body1, `id="login-form"`) != 1 {
		t.Errorf("first attempt: expected exactly 1 id=\"login-form\" in response, got %d",
			strings.Count(body1, `id="login-form"`))
	}

	// Второй неудачный логин — проверяем что структура та же (нет роста DOM)
	rec2 := doHTMXPost(t, e, "/login", form, csrfToken)
	if rec2.Code != http.StatusUnprocessableEntity {
		t.Fatalf("second attempt: expected 422, got %d", rec2.Code)
	}
	body2 := rec2.Body.String()
	if strings.Count(body2, `id="login-wrapper"`) != 1 {
		t.Errorf("second attempt: expected exactly 1 id=\"login-wrapper\" in response, got %d\nbody:\n%s",
			strings.Count(body2, `id="login-wrapper"`), body2)
	}
	if strings.Count(body2, `id="login-form"`) != 1 {
		t.Errorf("second attempt: expected exactly 1 id=\"login-form\" in response, got %d",
			strings.Count(body2, `id="login-form"`))
	}
}

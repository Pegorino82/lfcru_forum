//go:build integration

package admin_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/Pegorino82/lfcru_forum/internal/admin"
	"github.com/Pegorino82/lfcru_forum/internal/auth"
	appMiddleware "github.com/Pegorino82/lfcru_forum/internal/middleware"
	"github.com/Pegorino82/lfcru_forum/internal/news"
	"github.com/Pegorino82/lfcru_forum/internal/ratelimit"
	"github.com/Pegorino82/lfcru_forum/internal/session"
	"github.com/Pegorino82/lfcru_forum/internal/tmpl"
	"github.com/Pegorino82/lfcru_forum/internal/user"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"golang.org/x/crypto/bcrypt"
)

func newArticlesServer(t *testing.T, pool *pgxpool.Pool) *echo.Echo {
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

	newsRepo := news.NewRepo(pool)
	imagesRepo := admin.NewImagesRepo(pool)
	articlesHandler := admin.NewArticlesHandler(newsRepo, imagesRepo)

	e := echo.New()
	e.HideBanner = true
	e.Renderer = renderer
	e.Use(middleware.Recover())
	e.Use(appMiddleware.CSRFMiddleware())
	e.Use(auth.LoadSession(authSvc))

	adminGroup := e.Group("", admin.RequireAdminOrMod(renderer))
	adminGroup.GET("/admin/articles", articlesHandler.List)
	adminGroup.GET("/admin/articles/new", articlesHandler.New)
	adminGroup.POST("/admin/articles", articlesHandler.Create)
	adminGroup.GET("/admin/articles/:id/edit", articlesHandler.Edit)
	adminGroup.POST("/admin/articles/:id", articlesHandler.Update)
	adminGroup.GET("/admin/articles/:id/preview", articlesHandler.Preview)
	adminGroup.POST("/admin/articles/:id/status", articlesHandler.ChangeStatus)

	return e
}

func cleanArticlesTestData(t *testing.T, pool *pgxpool.Pool) {
	t.Helper()
	ctx := context.Background()
	pool.Exec(ctx, `DELETE FROM article_images WHERE article_id IN (SELECT id FROM news WHERE author_id IN (SELECT id FROM users WHERE email LIKE 'art-admintest%'))`)
	pool.Exec(ctx, `DELETE FROM news WHERE author_id IN (SELECT id FROM users WHERE email LIKE 'art-admintest%')`)
	pool.Exec(ctx, `DELETE FROM sessions WHERE user_id IN (SELECT id FROM users WHERE email LIKE 'art-admintest%')`)
	pool.Exec(ctx, `DELETE FROM users WHERE email LIKE 'art-admintest%'`)
}

func getArticlesCsrf(t *testing.T, e *echo.Echo, sessID string) string {
	t.Helper()
	rec := doGet(t, e, "/admin/articles", sessID)
	for _, c := range rec.Result().Cookies() {
		if c.Name == "_csrf" {
			return c.Value
		}
	}
	t.Fatal("no _csrf cookie in response")
	return ""
}

func doArticlesPost(t *testing.T, e *echo.Echo, path string, form url.Values, sessID, csrfToken string) *httptest.ResponseRecorder {
	t.Helper()
	form.Set("_csrf", csrfToken)
	req := httptest.NewRequest(http.MethodPost, path, strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Cookie", "session_id="+sessID+"; _csrf="+csrfToken)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	return rec
}

// SC-01: Admin открывает /admin/articles → 200 со списком.
func TestAdminArticles_List(t *testing.T) {
	pool := testDB(t)
	cleanArticlesTestData(t, pool)
	defer cleanArticlesTestData(t, pool)

	_, sessID := createUserWithRole(t, pool, "art-admintest-a1@test.com", "art_admintest_a1", "admin")
	e := newArticlesServer(t, pool)

	rec := doGet(t, e, "/admin/articles", sessID)
	if rec.Code != http.StatusOK {
		t.Fatalf("want 200, got %d\n%s", rec.Code, rec.Body.String())
	}
}

// SC-02: Admin создаёт статью → статья появляется в списке со статусом draft.
func TestAdminArticles_Create(t *testing.T) {
	pool := testDB(t)
	cleanArticlesTestData(t, pool)
	defer cleanArticlesTestData(t, pool)

	_, sessID := createUserWithRole(t, pool, "art-admintest-a2@test.com", "art_admintest_a2", "admin")
	e := newArticlesServer(t, pool)
	csrfToken := getArticlesCsrf(t, e, sessID)

	form := url.Values{
		"title":   {"art-admintest Тестовая статья"},
		"content": {"Содержимое в **Markdown**"},
	}
	rec := doArticlesPost(t, e, "/admin/articles", form, sessID, csrfToken)
	if rec.Code != http.StatusSeeOther {
		t.Fatalf("create: want 303, got %d\n%s", rec.Code, rec.Body.String())
	}

	// Verify draft exists in DB
	var status string
	err := pool.QueryRow(context.Background(),
		`SELECT status FROM news WHERE title = $1`, "art-admintest Тестовая статья",
	).Scan(&status)
	if err != nil {
		t.Fatalf("query created article: %v", err)
	}
	if status != "draft" {
		t.Errorf("expected status=draft, got %s", status)
	}
}

// EC-01: Черновик не виден в /news.
func TestAdminArticles_DraftNotVisibleInPublicNews(t *testing.T) {
	pool := testDB(t)
	cleanArticlesTestData(t, pool)
	defer cleanArticlesTestData(t, pool)

	authorID, sessID := createUserWithRole(t, pool, "art-admintest-a3@test.com", "art_admintest_a3", "admin")
	e := newArticlesServer(t, pool)
	csrfToken := getArticlesCsrf(t, e, sessID)

	form := url.Values{
		"title":   {"art-admintest EC-01 черновик"},
		"content": {"Содержимое"},
	}
	doArticlesPost(t, e, "/admin/articles", form, sessID, csrfToken)
	_ = authorID

	// Check that draft is NOT published
	var count int
	pool.QueryRow(context.Background(),
		`SELECT COUNT(*) FROM news WHERE title = 'art-admintest EC-01 черновик' AND status = 'published'`,
	).Scan(&count)
	if count != 0 {
		t.Error("draft should not be published")
	}
}

// SC-03: Admin редактирует черновик → изменения сохранены.
func TestAdminArticles_Edit(t *testing.T) {
	pool := testDB(t)
	cleanArticlesTestData(t, pool)
	defer cleanArticlesTestData(t, pool)

	authorID, sessID := createUserWithRole(t, pool, "art-admintest-a4@test.com", "art_admintest_a4", "admin")
	e := newArticlesServer(t, pool)

	// Insert article directly
	var articleID int64
	pool.QueryRow(context.Background(),
		`INSERT INTO news (title, content, status, author_id) VALUES ($1, $2, 'draft', $3) RETURNING id`,
		"art-admintest edit original", "original content", authorID,
	).Scan(&articleID)

	csrfToken := getArticlesCsrf(t, e, sessID)
	form := url.Values{
		"title":   {"art-admintest edit updated"},
		"content": {"updated content"},
	}
	rec := doArticlesPost(t, e, fmt.Sprintf("/admin/articles/%d", articleID), form, sessID, csrfToken)
	if rec.Code != http.StatusSeeOther {
		t.Fatalf("update: want 303, got %d\n%s", rec.Code, rec.Body.String())
	}

	var title string
	pool.QueryRow(context.Background(), `SELECT title FROM news WHERE id = $1`, articleID).Scan(&title)
	if title != "art-admintest edit updated" {
		t.Errorf("title not updated, got %q", title)
	}
}

// SC-04: Admin открывает превью черновика → 200.
func TestAdminArticles_Preview(t *testing.T) {
	pool := testDB(t)
	cleanArticlesTestData(t, pool)
	defer cleanArticlesTestData(t, pool)

	authorID, sessID := createUserWithRole(t, pool, "art-admintest-a5@test.com", "art_admintest_a5", "admin")
	e := newArticlesServer(t, pool)

	var articleID int64
	pool.QueryRow(context.Background(),
		`INSERT INTO news (title, content, status, author_id) VALUES ($1, $2, 'draft', $3) RETURNING id`,
		"art-admintest preview article", "**Bold text**", authorID,
	).Scan(&articleID)

	rec := doGet(t, e, fmt.Sprintf("/admin/articles/%d/preview", articleID), sessID)
	if rec.Code != http.StatusOK {
		t.Fatalf("preview: want 200, got %d\n%s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), "art-admintest preview article") {
		t.Error("expected article title in preview")
	}
}

// SC-05: Admin меняет статус черновика на in_review.
func TestAdminArticles_StatusToInReview(t *testing.T) {
	pool := testDB(t)
	cleanArticlesTestData(t, pool)
	defer cleanArticlesTestData(t, pool)

	authorID, sessID := createUserWithRole(t, pool, "art-admintest-a6@test.com", "art_admintest_a6", "admin")
	e := newArticlesServer(t, pool)

	var articleID int64
	pool.QueryRow(context.Background(),
		`INSERT INTO news (title, content, status, author_id) VALUES ($1, $2, 'draft', $3) RETURNING id`,
		"art-admintest status article", "Some content", authorID,
	).Scan(&articleID)

	csrfToken := getArticlesCsrf(t, e, sessID)
	form := url.Values{"status": {"in_review"}}
	rec := doArticlesPost(t, e, fmt.Sprintf("/admin/articles/%d/status", articleID), form, sessID, csrfToken)
	if rec.Code != http.StatusSeeOther {
		t.Fatalf("status change: want 303, got %d\n%s", rec.Code, rec.Body.String())
	}

	var status string
	pool.QueryRow(context.Background(), `SELECT status FROM news WHERE id = $1`, articleID).Scan(&status)
	if status != "in_review" {
		t.Errorf("expected in_review, got %s", status)
	}
}

// SC-06 / EC-03: Публикация → статус published, статья видна по /news ID.
func TestAdminArticles_Publish(t *testing.T) {
	pool := testDB(t)
	cleanArticlesTestData(t, pool)
	defer cleanArticlesTestData(t, pool)

	authorID, sessID := createUserWithRole(t, pool, "art-admintest-a7@test.com", "art_admintest_a7", "admin")
	e := newArticlesServer(t, pool)

	var articleID int64
	pool.QueryRow(context.Background(),
		`INSERT INTO news (title, content, status, author_id) VALUES ($1, $2, 'draft', $3) RETURNING id`,
		"art-admintest publish article", "Full content", authorID,
	).Scan(&articleID)

	csrfToken := getArticlesCsrf(t, e, sessID)
	form := url.Values{"status": {"published"}}
	rec := doArticlesPost(t, e, fmt.Sprintf("/admin/articles/%d/status", articleID), form, sessID, csrfToken)
	if rec.Code != http.StatusSeeOther {
		t.Fatalf("publish: want 303, got %d\n%s", rec.Code, rec.Body.String())
	}

	var status string
	var publishedAt *time.Time
	pool.QueryRow(context.Background(),
		`SELECT status, published_at FROM news WHERE id = $1`, articleID,
	).Scan(&status, &publishedAt)
	if status != "published" {
		t.Errorf("expected published, got %s", status)
	}
	if publishedAt == nil {
		t.Error("published_at should be set after publishing")
	}
}

// EC-05: Недопустимый переход статуса → 400.
func TestAdminArticles_InvalidStatusTransition(t *testing.T) {
	pool := testDB(t)
	cleanArticlesTestData(t, pool)
	defer cleanArticlesTestData(t, pool)

	authorID, sessID := createUserWithRole(t, pool, "art-admintest-a8@test.com", "art_admintest_a8", "admin")
	e := newArticlesServer(t, pool)

	// published → in_review is invalid
	var articleID int64
	pool.QueryRow(context.Background(),
		`INSERT INTO news (title, content, status, author_id, published_at) VALUES ($1, $2, 'published', $3, now()) RETURNING id`,
		"art-admintest invalid transition", "content", authorID,
	).Scan(&articleID)

	csrfToken := getArticlesCsrf(t, e, sessID)
	form := url.Values{"status": {"in_review"}}
	rec := doArticlesPost(t, e, fmt.Sprintf("/admin/articles/%d/status", articleID), form, sessID, csrfToken)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("want 400 for invalid transition, got %d\n%s", rec.Code, rec.Body.String())
	}
}

// FT-015 Regression: превью содержит баннер "Режим превью" и ссылку "Назад к редактору".
func TestAdminArticles_Preview_HasPreviewBanner(t *testing.T) {
	pool := testDB(t)
	cleanArticlesTestData(t, pool)
	defer cleanArticlesTestData(t, pool)

	authorID, sessID := createUserWithRole(t, pool, "art-admintest-b1@test.com", "art_admintest_b1", "admin")
	e := newArticlesServer(t, pool)

	var articleID int64
	pool.QueryRow(context.Background(),
		`INSERT INTO news (title, content, status, author_id) VALUES ($1, $2, 'draft', $3) RETURNING id`,
		"art-admintest banner article", "content", authorID,
	).Scan(&articleID)

	rec := doGet(t, e, fmt.Sprintf("/admin/articles/%d/preview", articleID), sessID)
	if rec.Code != http.StatusOK {
		t.Fatalf("preview: want 200, got %d\n%s", rec.Code, rec.Body.String())
	}
	body := rec.Body.String()
	if !strings.Contains(body, "Режим превью") {
		t.Error("preview page must contain 'Режим превью' banner")
	}
	if !strings.Contains(body, "Назад к редактору") {
		t.Error("preview page must contain 'Назад к редактору' link")
	}
	if strings.Contains(body, "Оставить комментарий") || strings.Contains(body, "hx-post") {
		t.Error("preview page must not contain comment form")
	}
}

// FT-015 Regression: после сохранения статьи редирект содержит ?saved=1, страница показывает подтверждение.
func TestAdminArticles_Update_ShowsSavedConfirmation(t *testing.T) {
	pool := testDB(t)
	cleanArticlesTestData(t, pool)
	defer cleanArticlesTestData(t, pool)

	authorID, sessID := createUserWithRole(t, pool, "art-admintest-b2@test.com", "art_admintest_b2", "admin")
	e := newArticlesServer(t, pool)

	var articleID int64
	pool.QueryRow(context.Background(),
		`INSERT INTO news (title, content, status, author_id) VALUES ($1, $2, 'draft', $3) RETURNING id`,
		"art-admintest save confirm", "original", authorID,
	).Scan(&articleID)

	csrfToken := getArticlesCsrf(t, e, sessID)
	form := url.Values{
		"title":   {"art-admintest save confirm updated"},
		"content": {"updated"},
	}
	rec := doArticlesPost(t, e, fmt.Sprintf("/admin/articles/%d", articleID), form, sessID, csrfToken)
	if rec.Code != http.StatusSeeOther {
		t.Fatalf("update: want 303, got %d\n%s", rec.Code, rec.Body.String())
	}
	location := rec.Header().Get("Location")
	if !strings.Contains(location, "saved=1") {
		t.Errorf("redirect location must contain saved=1, got %q", location)
	}

	// Follow redirect — edit page must show saved confirmation.
	rec2 := doGet(t, e, location, sessID)
	if rec2.Code != http.StatusOK {
		t.Fatalf("edit after save: want 200, got %d\n%s", rec2.Code, rec2.Body.String())
	}
	if !strings.Contains(rec2.Body.String(), "Статья сохранена") {
		t.Error("edit page must show 'Статья сохранена' confirmation after save")
	}
}

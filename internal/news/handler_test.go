//go:build integration

package news_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/Pegorino82/lfcru_forum/internal/auth"
	"github.com/Pegorino82/lfcru_forum/internal/comment"
	appMiddleware "github.com/Pegorino82/lfcru_forum/internal/middleware"
	"github.com/Pegorino82/lfcru_forum/internal/news"
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

// cleanArticleData removes test data related to news and comments.
func cleanArticleData(t *testing.T, pool *pgxpool.Pool) {
	t.Helper()
	ctx := context.Background()
	pool.Exec(ctx, `DELETE FROM news_comments WHERE news_id IN (SELECT id FROM news WHERE title LIKE 'test-%')`)
	pool.Exec(ctx, `DELETE FROM news WHERE title LIKE 'test-%'`)
	pool.Exec(ctx, `DELETE FROM sessions WHERE user_id IN (SELECT id FROM users WHERE email LIKE 'newstest%')`)
	pool.Exec(ctx, `DELETE FROM users WHERE email LIKE 'newstest%'`)
}

// newTestServer builds an Echo server with news routes registered.
func newTestServer(t *testing.T, pool *pgxpool.Pool) *echo.Echo {
	t.Helper()
	renderer, err := tmpl.New(os.DirFS(templatesPath), "templates/")
	if err != nil {
		t.Fatalf("load templates: %v", err)
	}

	userRepo := user.NewRepo(pool)
	svc := auth.NewService(
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

	commentRepo := comment.NewRepo(pool)
	commentSvc := comment.NewService(commentRepo, userRepo)
	newsRepo := news.NewRepo(pool)

	e := echo.New()
	e.HideBanner = true
	e.Renderer = renderer
	e.Use(middleware.Recover())
	e.Use(appMiddleware.CSRFMiddleware())
	e.Use(auth.LoadSession(svc))
	news.NewHandler(newsRepo, commentRepo, commentSvc).RegisterRoutes(e)

	return e
}

// createUser inserts a test user and returns (userID, sessionCookie).
func createUser(t *testing.T, pool *pgxpool.Pool, email, username string) (int64, string) {
	t.Helper()
	ctx := context.Background()
	hash, _ := bcrypt.GenerateFromPassword([]byte("password"), bcrypt.MinCost)
	var userID int64
	if err := pool.QueryRow(ctx,
		`INSERT INTO users (email, username, pass_hash) VALUES ($1, $2, $3) RETURNING id`,
		email, username, hash,
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

// insertNews inserts a news article and returns its ID.
func insertNews(t *testing.T, pool *pgxpool.Pool, title string, published bool, authorID int64) int64 {
	t.Helper()
	ctx := context.Background()
	var id int64
	status := "draft"
	var publishedAt interface{}
	if published {
		status = "published"
		publishedAt = time.Now()
	}
	if err := pool.QueryRow(ctx,
		`INSERT INTO news (title, content, status, author_id, published_at)
		 VALUES ($1, $2, $3::news_status, $4, $5) RETURNING id`,
		title, "Содержимое статьи "+title, status, authorID, publishedAt,
	).Scan(&id); err != nil {
		t.Fatalf("insert news: %v", err)
	}
	return id
}

// ─── helpers ─────────────────────────────────────────────────────────────────

func doGet(t *testing.T, e *echo.Echo, path string, sessID string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(http.MethodGet, path, nil)
	if sessID != "" {
		req.Header.Set("Cookie", "session_id="+sessID)
	}
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	return rec
}

// getCsrfToken extracts _csrf token value from Set-Cookie response headers.
func getCsrfToken(rec *httptest.ResponseRecorder) string {
	for _, c := range rec.Result().Cookies() {
		if c.Name == "_csrf" {
			return c.Value
		}
	}
	return ""
}

// doPost makes a POST request with proper session + CSRF cookies and form body.
func doPost(t *testing.T, e *echo.Echo, path string, form url.Values, sessID, csrfToken string, htmx bool) *httptest.ResponseRecorder {
	t.Helper()
	// CSRF token must appear in both form field and cookie.
	if csrfToken != "" {
		form.Set("_csrf", csrfToken)
	}
	req := httptest.NewRequest(http.MethodPost, path, strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	var cookies []string
	if sessID != "" {
		cookies = append(cookies, "session_id="+sessID)
	}
	if csrfToken != "" {
		cookies = append(cookies, "_csrf="+csrfToken)
	}
	if len(cookies) > 0 {
		req.Header.Set("Cookie", strings.Join(cookies, "; "))
	}
	if htmx {
		req.Header.Set("HX-Request", "true")
	}
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	return rec
}

// ─── Tests ────────────────────────────────────────────────────────────────────

// ─── News list tests ──────────────────────────────────────────────────────────

func TestShowList_OK(t *testing.T) {
	pool := testDB(t)
	cleanArticleData(t, pool)
	defer cleanArticleData(t, pool)

	authorID, _ := createUser(t, pool, "newstest-list1@test.com", "newstest-list1")
	insertNews(t, pool, "test-list-article-1", true, authorID)
	insertNews(t, pool, "test-list-article-2", true, authorID)

	e := newTestServer(t, pool)
	rec := doGet(t, e, "/news", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	body := rec.Body.String()
	if !strings.Contains(body, "test-list-article-1") {
		t.Error("expected article-1 in list")
	}
	if !strings.Contains(body, "test-list-article-2") {
		t.Error("expected article-2 in list")
	}
}

func TestShowList_NoDrafts(t *testing.T) {
	pool := testDB(t)
	cleanArticleData(t, pool)
	defer cleanArticleData(t, pool)

	authorID, _ := createUser(t, pool, "newstest-listdraft@test.com", "newstest-listdraft")
	insertNews(t, pool, "test-list-published-ok", true, authorID)
	insertNews(t, pool, "test-list-draft-hidden", false, authorID)

	e := newTestServer(t, pool)
	rec := doGet(t, e, "/news", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	if strings.Contains(rec.Body.String(), "test-list-draft-hidden") {
		t.Error("draft should not appear in list")
	}
}

func TestShowList_InvalidPage(t *testing.T) {
	pool := testDB(t)
	cleanArticleData(t, pool)
	defer cleanArticleData(t, pool)

	e := newTestServer(t, pool)
	for _, q := range []string{"?page=0", "?page=-1", "?page=abc", ""} {
		rec := doGet(t, e, "/news"+q, "")
		if rec.Code != http.StatusOK {
			t.Errorf("query %q: expected 200, got %d", q, rec.Code)
		}
	}
}

func TestShowList_HTMXPartial(t *testing.T) {
	pool := testDB(t)
	cleanArticleData(t, pool)
	defer cleanArticleData(t, pool)

	e := newTestServer(t, pool)
	req := httptest.NewRequest(http.MethodGet, "/news", nil)
	req.Header.Set("HX-Request", "true")
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	if strings.Contains(rec.Body.String(), "<html") {
		t.Error("HTMX partial should not contain <html>")
	}
}

// ─── Article tests ────────────────────────────────────────────────────────────

func TestShowArticle_InvalidID(t *testing.T) {
	pool := testDB(t)
	cleanArticleData(t, pool)
	defer cleanArticleData(t, pool)

	e := newTestServer(t, pool)
	rec := doGet(t, e, "/news/abc", "")
	if rec.Code != http.StatusNotFound {
		t.Errorf("expected 404 for non-numeric id, got %d", rec.Code)
	}
}

func TestShowArticle_NotFound(t *testing.T) {
	pool := testDB(t)
	cleanArticleData(t, pool)
	defer cleanArticleData(t, pool)

	e := newTestServer(t, pool)
	rec := doGet(t, e, "/news/999999", "")
	if rec.Code != http.StatusNotFound {
		t.Errorf("expected 404 for missing article, got %d", rec.Code)
	}
}

func TestShowArticle_UnpublishedIsNotFound(t *testing.T) {
	pool := testDB(t)
	cleanArticleData(t, pool)
	defer cleanArticleData(t, pool)

	authorID, _ := createUser(t, pool, "newstest-author@test.com", "newstest-author")
	id := insertNews(t, pool, "test-draft", false, authorID)

	e := newTestServer(t, pool)
	rec := doGet(t, e, fmt.Sprintf("/news/%d", id), "")
	if rec.Code != http.StatusNotFound {
		t.Errorf("expected 404 for unpublished article, got %d", rec.Code)
	}
}

func TestShowArticle_Published(t *testing.T) {
	pool := testDB(t)
	cleanArticleData(t, pool)
	defer cleanArticleData(t, pool)

	authorID, _ := createUser(t, pool, "newstest-pub@test.com", "newstest-pub")
	id := insertNews(t, pool, "test-Liverpool победил", true, authorID)

	e := newTestServer(t, pool)
	rec := doGet(t, e, fmt.Sprintf("/news/%d", id), "")
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "test-Liverpool победил") {
		t.Error("expected article title in body")
	}
}

func TestShowArticle_GuestAccess(t *testing.T) {
	pool := testDB(t)
	cleanArticleData(t, pool)
	defer cleanArticleData(t, pool)

	authorID, _ := createUser(t, pool, "newstest-guest@test.com", "newstest-guest")
	id := insertNews(t, pool, "test-guest article", true, authorID)

	e := newTestServer(t, pool)
	rec := doGet(t, e, fmt.Sprintf("/news/%d", id), "")
	if rec.Code != http.StatusOK {
		t.Fatalf("guest should see 200, got %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "Войдите") {
		t.Error("guest should see login prompt in comments")
	}
}

func TestShowArticle_EmptyComments_Guest(t *testing.T) {
	pool := testDB(t)
	cleanArticleData(t, pool)
	defer cleanArticleData(t, pool)

	authorID, _ := createUser(t, pool, "newstest-ec1@test.com", "newstest-ec1")
	id := insertNews(t, pool, "test-no-comments", true, authorID)

	e := newTestServer(t, pool)
	rec := doGet(t, e, fmt.Sprintf("/news/%d", id), "")
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "Войдите") {
		t.Error("expected guest login prompt in empty-state message")
	}
}

func TestShowArticle_EmptyComments_AuthUser(t *testing.T) {
	pool := testDB(t)
	cleanArticleData(t, pool)
	defer cleanArticleData(t, pool)

	authorID, sessID := createUser(t, pool, "newstest-ec2@test.com", "newstest-ec2")
	id := insertNews(t, pool, "test-no-comments-auth", true, authorID)

	e := newTestServer(t, pool)
	rec := doGet(t, e, fmt.Sprintf("/news/%d", id), sessID)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "Будьте первым") {
		t.Error("expected 'Будьте первым' in empty-state for auth user")
	}
}

func TestShowArticle_HTMXPartial(t *testing.T) {
	pool := testDB(t)
	cleanArticleData(t, pool)
	defer cleanArticleData(t, pool)

	authorID, _ := createUser(t, pool, "newstest-htmx@test.com", "newstest-htmx")
	id := insertNews(t, pool, "test-htmx-article", true, authorID)

	e := newTestServer(t, pool)
	req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/news/%d", id), nil)
	req.Header.Set("HX-Request", "true")
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	if strings.Contains(rec.Body.String(), "<html") {
		t.Error("HTMX partial should not contain <html>")
	}
}

func TestCreateComment_GuestRedirectsToLogin(t *testing.T) {
	pool := testDB(t)
	cleanArticleData(t, pool)
	defer cleanArticleData(t, pool)

	authorID, _ := createUser(t, pool, "newstest-gcomment@test.com", "newstest-gcomment")
	id := insertNews(t, pool, "test-guest-comment", true, authorID)

	e := newTestServer(t, pool)
	getRec := doGet(t, e, fmt.Sprintf("/news/%d", id), "")
	csrfToken := getCsrfToken(getRec)

	form := url.Values{"content": {"Привет!"}}
	rec := doPost(t, e, fmt.Sprintf("/news/%d/comments", id), form, "", csrfToken, false)

	if rec.Code != http.StatusFound {
		t.Errorf("guest should be redirected (302), got %d", rec.Code)
	}
	if !strings.Contains(rec.Header().Get("Location"), "/login") {
		t.Errorf("expected redirect to /login, got %s", rec.Header().Get("Location"))
	}
}

func TestCreateComment_EmptyContent(t *testing.T) {
	pool := testDB(t)
	cleanArticleData(t, pool)
	defer cleanArticleData(t, pool)

	authorID, sessID := createUser(t, pool, "newstest-empty@test.com", "newstest-empty")
	id := insertNews(t, pool, "test-empty-comment", true, authorID)

	e := newTestServer(t, pool)
	getRec := doGet(t, e, fmt.Sprintf("/news/%d", id), sessID)
	csrfToken := getCsrfToken(getRec)

	form := url.Values{"content": {""}}
	rec := doPost(t, e, fmt.Sprintf("/news/%d/comments", id), form, sessID, csrfToken, true)

	if rec.Code != http.StatusUnprocessableEntity {
		t.Errorf("expected 422, got %d\nbody: %s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), "пустым") {
		t.Errorf("expected empty-content error, got: %s", rec.Body.String())
	}
}

func TestCreateComment_TooLong(t *testing.T) {
	pool := testDB(t)
	cleanArticleData(t, pool)
	defer cleanArticleData(t, pool)

	authorID, sessID := createUser(t, pool, "newstest-long@test.com", "newstest-long")
	id := insertNews(t, pool, "test-long-comment", true, authorID)

	e := newTestServer(t, pool)
	getRec := doGet(t, e, fmt.Sprintf("/news/%d", id), sessID)
	csrfToken := getCsrfToken(getRec)

	form := url.Values{"content": {strings.Repeat("а", 10001)}}
	rec := doPost(t, e, fmt.Sprintf("/news/%d/comments", id), form, sessID, csrfToken, true)

	if rec.Code != http.StatusUnprocessableEntity {
		t.Errorf("expected 422 for too-long comment, got %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "10 000") {
		t.Errorf("expected max-length error, got: %s", rec.Body.String())
	}
}

func TestCreateComment_ValidHTMX(t *testing.T) {
	pool := testDB(t)
	cleanArticleData(t, pool)
	defer cleanArticleData(t, pool)

	authorID, sessID := createUser(t, pool, "newstest-valid@test.com", "newstest-valid")
	id := insertNews(t, pool, "test-valid-comment", true, authorID)

	e := newTestServer(t, pool)
	getRec := doGet(t, e, fmt.Sprintf("/news/%d", id), sessID)
	csrfToken := getCsrfToken(getRec)

	form := url.Values{"content": {"Отличная статья!"}}
	rec := doPost(t, e, fmt.Sprintf("/news/%d/comments", id), form, sessID, csrfToken, true)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d\nbody: %s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), "Отличная статья!") {
		t.Error("expected new comment content in response")
	}
	if strings.Contains(rec.Body.String(), "<html") {
		t.Error("HTMX response should not be a full page")
	}
}

func TestCreateComment_ValidNonHTMX_Redirect(t *testing.T) {
	pool := testDB(t)
	cleanArticleData(t, pool)
	defer cleanArticleData(t, pool)

	authorID, sessID := createUser(t, pool, "newstest-redirect@test.com", "newstest-redirect")
	id := insertNews(t, pool, "test-redirect-comment", true, authorID)

	e := newTestServer(t, pool)
	getRec := doGet(t, e, fmt.Sprintf("/news/%d", id), sessID)
	csrfToken := getCsrfToken(getRec)

	form := url.Values{"content": {"Хороший материал"}}
	rec := doPost(t, e, fmt.Sprintf("/news/%d/comments", id), form, sessID, csrfToken, false)

	if rec.Code != http.StatusSeeOther {
		t.Errorf("expected 303 redirect, got %d", rec.Code)
	}
	expected := fmt.Sprintf("/news/%d", id)
	if !strings.Contains(rec.Header().Get("Location"), expected) {
		t.Errorf("expected redirect to %s, got %s", expected, rec.Header().Get("Location"))
	}
}

func TestCreateComment_Reply(t *testing.T) {
	pool := testDB(t)
	cleanArticleData(t, pool)
	defer cleanArticleData(t, pool)

	authorID, sessID := createUser(t, pool, "newstest-reply@test.com", "newstest-reply")
	id := insertNews(t, pool, "test-reply-comment", true, authorID)

	var parentID int64
	pool.QueryRow(context.Background(),
		`INSERT INTO news_comments (news_id, author_id, content) VALUES ($1, $2, $3) RETURNING id`,
		id, authorID, "Родительский комментарий",
	).Scan(&parentID)

	e := newTestServer(t, pool)
	getRec := doGet(t, e, fmt.Sprintf("/news/%d", id), sessID)
	csrfToken := getCsrfToken(getRec)

	form := url.Values{
		"content":   {"Ответ на комментарий"},
		"parent_id": {fmt.Sprintf("%d", parentID)},
	}
	rec := doPost(t, e, fmt.Sprintf("/news/%d/comments", id), form, sessID, csrfToken, true)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d\nbody: %s", rec.Code, rec.Body.String())
	}
	body := rec.Body.String()
	if !strings.Contains(body, "Ответ на комментарий") {
		t.Error("expected reply content in response")
	}
	if !strings.Contains(body, "Родительский комментарий") {
		t.Error("expected parent snippet (quote) in response")
	}
}

func TestCreateComment_ReplyParentNotFound(t *testing.T) {
	pool := testDB(t)
	cleanArticleData(t, pool)
	defer cleanArticleData(t, pool)

	authorID, sessID := createUser(t, pool, "newstest-noreply@test.com", "newstest-noreply")
	id := insertNews(t, pool, "test-noreply", true, authorID)

	e := newTestServer(t, pool)
	getRec := doGet(t, e, fmt.Sprintf("/news/%d", id), sessID)
	csrfToken := getCsrfToken(getRec)

	form := url.Values{
		"content":   {"Ответ на несуществующий"},
		"parent_id": {"999999"},
	}
	rec := doPost(t, e, fmt.Sprintf("/news/%d/comments", id), form, sessID, csrfToken, true)

	if rec.Code != http.StatusUnprocessableEntity {
		t.Errorf("expected 422, got %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "не найден") {
		t.Errorf("expected parent-not-found error, got: %s", rec.Body.String())
	}
}

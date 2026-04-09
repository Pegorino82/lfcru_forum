//go:build integration

package forum_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/Pegorino82/lfcru_forum/internal/auth"
	"github.com/Pegorino82/lfcru_forum/internal/forum"
	"github.com/Pegorino82/lfcru_forum/internal/ratelimit"
	appMiddleware "github.com/Pegorino82/lfcru_forum/internal/middleware"
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
	templatesPath = "../../templates"
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
		if err := goose.Up(db, "../../migrations"); err != nil {
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

// cleanForumData removes test data from forum tables.
func cleanForumData(t *testing.T, pool *pgxpool.Pool) {
	t.Helper()
	ctx := context.Background()
	pool.Exec(ctx, `DELETE FROM forum_posts WHERE topic_id IN (SELECT id FROM forum_topics WHERE title LIKE 'test-%')`)
	pool.Exec(ctx, `DELETE FROM forum_topics WHERE title LIKE 'test-%'`)
	pool.Exec(ctx, `DELETE FROM forum_sections WHERE title LIKE 'test-%'`)
	pool.Exec(ctx, `DELETE FROM sessions WHERE user_id IN (SELECT id FROM users WHERE email LIKE 'forumtest%')`)
	pool.Exec(ctx, `DELETE FROM users WHERE email LIKE 'forumtest%'`)
}

// newTestServer builds an Echo server with forum routes registered.
func newTestServer(t *testing.T, pool *pgxpool.Pool) *echo.Echo {
	t.Helper()
	renderer, err := tmpl.New(os.DirFS(templatesPath), "templates/")
	if err != nil {
		t.Fatalf("load templates: %v", err)
	}

	userRepo := user.NewRepo(pool)
	authCfg := auth.Config{
		BcryptCost:         bcrypt.MinCost,
		SessionLifetime:    30 * 24 * time.Hour,
		RateLimitWindow:    10 * time.Minute,
		RateLimitMax:       10,
		SessionGracePeriod: 5 * time.Minute,
		MaxSessionsPerUser: 10,
	}
	authSvc := auth.NewService(
		userRepo,
		session.NewRepo(pool),
		ratelimit.NewLoginAttemptRepo(pool),
		authCfg,
	)

	forumRepo := forum.NewRepo(pool)
	forumSvc := forum.NewService(forumRepo)

	e := echo.New()
	e.HideBanner = true
	e.Renderer = renderer
	e.Use(middleware.Recover())
	e.Use(appMiddleware.CSRFMiddleware())
	e.Use(auth.LoadSession(authSvc))

	// Register forum routes
	forumHandler := forum.NewHandler(forumSvc, renderer)

	modGroup := e.Group("", auth.RequireAuth, auth.RequireRole(renderer, "moderator", "admin"))
	modGroup.GET("/forum/sections/new", forumHandler.NewSection)
	modGroup.POST("/forum/sections", forumHandler.CreateSection)
	modGroup.GET("/forum/sections/:id/topics/new", forumHandler.NewTopic)
	modGroup.POST("/forum/sections/:id/topics", forumHandler.CreateTopic)

	e.GET("/forum", forumHandler.Index)
	e.GET("/forum/sections/:id", forumHandler.ShowSection)
	e.GET("/forum/topics/:id", forumHandler.ShowTopic)

	authGroup := e.Group("", auth.RequireAuth)
	authGroup.POST("/forum/topics/:id/posts", forumHandler.CreatePost)

	return e
}

// createUser inserts a test user and returns (userID, sessionCookie).
// role defaults to "user" if empty.
func createUser(t *testing.T, pool *pgxpool.Pool, email, username, role string) (int64, string) {
	t.Helper()
	ctx := context.Background()
	hash, _ := bcrypt.GenerateFromPassword([]byte("password"), bcrypt.MinCost)

	if role == "" {
		role = "user"
	}

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

// newSection inserts a forum section and returns its ID.
func newSection(t *testing.T, pool *pgxpool.Pool, title, description string, sortOrder int) int64 {
	t.Helper()
	ctx := context.Background()
	var id int64
	if err := pool.QueryRow(ctx,
		`INSERT INTO forum_sections (title, description, sort_order) VALUES ($1, $2, $3) RETURNING id`,
		title, description, sortOrder,
	).Scan(&id); err != nil {
		t.Fatalf("insert section: %v", err)
	}
	return id
}

// newTopic inserts a forum topic and returns its ID.
func newTopic(t *testing.T, pool *pgxpool.Pool, sectionID int64, authorID int64, title string) int64 {
	t.Helper()
	ctx := context.Background()
	var id int64
	if err := pool.QueryRow(ctx,
		`INSERT INTO forum_topics (section_id, title, author_id) VALUES ($1, $2, $3) RETURNING id`,
		sectionID, title, authorID,
	).Scan(&id); err != nil {
		t.Fatalf("insert topic: %v", err)
	}
	return id
}

// ─── helpers ─────────────────────────────────────────────────────────────────

func insertTestSection(t *testing.T, pool *pgxpool.Pool, title, description string, sortOrder int) int64 {
	t.Helper()
	ctx := context.Background()
	var id int64
	if err := pool.QueryRow(ctx,
		`INSERT INTO forum_sections (title, description, sort_order) VALUES ($1, $2, $3) RETURNING id`,
		title, description, sortOrder,
	).Scan(&id); err != nil {
		t.Fatalf("insert section: %v", err)
	}
	return id
}

func insertTestTopic(t *testing.T, pool *pgxpool.Pool, sectionID, authorID int64, title string) int64 {
	t.Helper()
	ctx := context.Background()
	var id int64
	if err := pool.QueryRow(ctx,
		`INSERT INTO forum_topics (section_id, author_id, title) VALUES ($1, $2, $3) RETURNING id`,
		sectionID, authorID, title,
	).Scan(&id); err != nil {
		t.Fatalf("insert topic: %v", err)
	}
	return id
}

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
	req := httptest.NewRequest(http.MethodPost, path, strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	if sessID != "" {
		req.Header.Set("Cookie", "session_id="+sessID)
	}
	if csrfToken != "" {
		req.Header.Set("Cookie", req.Header.Get("Cookie")+"; _csrf="+csrfToken)
	}
	if htmx {
		req.Header.Set("HX-Request", "true")
	}
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	return rec
}

// ─── Tests ───────────────────────────────────────────────────────────────────

func TestIndex_Guest(t *testing.T) {
	pool := testDB(t)
	cleanForumData(t, pool)
	defer cleanForumData(t, pool)

	e := newTestServer(t, pool)
	rec := doGet(t, e, "/forum", "")
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "Форум") {
		t.Error("expected 'Форум' in body")
	}
}

func TestIndex_EmptyList(t *testing.T) {
	pool := testDB(t)
	cleanForumData(t, pool)
	defer cleanForumData(t, pool)

	e := newTestServer(t, pool)
	rec := doGet(t, e, "/forum", "")
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "пока не созданы") {
		t.Error("expected empty state text")
	}
}

func TestShowSection_Exists(t *testing.T) {
	pool := testDB(t)
	cleanForumData(t, pool)
	defer cleanForumData(t, pool)

	sectionID := insertTestSection(t, pool, "test-section", "Test description", 0)

	e := newTestServer(t, pool)
	rec := doGet(t, e, "/forum/sections/"+strconv.FormatInt(sectionID, 10), "")
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "test-section") {
		t.Error("expected section title in body")
	}
}

func TestShowSection_NotFound(t *testing.T) {
	pool := testDB(t)
	cleanForumData(t, pool)
	defer cleanForumData(t, pool)

	e := newTestServer(t, pool)
	rec := doGet(t, e, "/forum/sections/999", "")
	if rec.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", rec.Code)
	}
}

func TestShowSection_InvalidID(t *testing.T) {
	pool := testDB(t)
	e := newTestServer(t, pool)
	rec := doGet(t, e, "/forum/sections/abc", "")
	if rec.Code != http.StatusNotFound {
		t.Errorf("expected 404 for invalid id, got %d", rec.Code)
	}
}

func TestShowTopic_Exists(t *testing.T) {
	pool := testDB(t)
	cleanForumData(t, pool)
	defer cleanForumData(t, pool)

	authorID, _ := createUser(t, pool, "forumtest-author@test.com", "author", "user")
	sectionID := insertTestSection(t, pool, "test-section", "", 0)
	topicID := insertTestTopic(t, pool, sectionID, authorID, "test-topic")

	e := newTestServer(t, pool)
	rec := doGet(t, e, "/forum/topics/"+strconv.FormatInt(topicID, 10), "")
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "test-topic") {
		t.Error("expected topic title in body")
	}
}

func TestShowTopic_NotFound(t *testing.T) {
	pool := testDB(t)
	e := newTestServer(t, pool)
	rec := doGet(t, e, "/forum/topics/999", "")
	if rec.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", rec.Code)
	}
}

func TestCreatePost_User(t *testing.T) {
	pool := testDB(t)
	cleanForumData(t, pool)
	defer cleanForumData(t, pool)

	userID, sessID := createUser(t, pool, "forumtest-user@test.com", "user", "user")
	sectionID := insertTestSection(t, pool, "test-section", "", 0)
	topicID := insertTestTopic(t, pool, sectionID, userID, "test-topic")

	e := newTestServer(t, pool)

	// Get CSRF token
	getReq := httptest.NewRequest(http.MethodGet, "/forum/topics/"+strconv.FormatInt(topicID, 10), nil)
	getReq.Header.Set("Cookie", "session_id="+sessID)
	getRec := httptest.NewRecorder()
	e.ServeHTTP(getRec, getReq)
	csrfToken := getCsrfToken(getRec)

	// Post
	form := url.Values{
		"content": {"test content"},
		"_csrf":   {csrfToken},
	}
	rec := doPost(t, e, "/forum/topics/"+strconv.FormatInt(topicID, 10)+"/posts", form, sessID, csrfToken, true)
	if rec.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d", rec.Code)
	}
}

func TestCreatePost_Guest(t *testing.T) {
	pool := testDB(t)
	cleanForumData(t, pool)
	defer cleanForumData(t, pool)

	authorID, _ := createUser(t, pool, "forumtest-author@test.com", "author", "user")
	sectionID := insertTestSection(t, pool, "test-section", "", 0)
	topicID := insertTestTopic(t, pool, sectionID, authorID, "test-topic")

	e := newTestServer(t, pool)

	// Get CSRF token
	getReq := httptest.NewRequest(http.MethodGet, "/forum/topics/"+strconv.FormatInt(topicID, 10), nil)
	getRec := httptest.NewRecorder()
	e.ServeHTTP(getRec, getReq)
	csrfToken := getCsrfToken(getRec)

	// Post (guest)
	form := url.Values{
		"content": {"test content"},
		"_csrf":   {csrfToken},
	}
	req := httptest.NewRequest(http.MethodPost, "/forum/topics/"+strconv.FormatInt(topicID, 10)+"/posts", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Cookie", "_csrf="+csrfToken)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	if rec.Code != http.StatusFound {
		t.Errorf("expected 302 redirect to /login, got %d", rec.Code)
	}
}

func TestCreatePost_EmptyContent(t *testing.T) {
	pool := testDB(t)
	cleanForumData(t, pool)
	defer cleanForumData(t, pool)

	userID, sessID := createUser(t, pool, "forumtest-user@test.com", "user", "user")
	sectionID := insertTestSection(t, pool, "test-section", "", 0)
	topicID := insertTestTopic(t, pool, sectionID, userID, "test-topic")

	e := newTestServer(t, pool)

	// Get CSRF token
	getReq := httptest.NewRequest(http.MethodGet, "/forum/topics/"+strconv.FormatInt(topicID, 10), nil)
	getReq.Header.Set("Cookie", "session_id="+sessID)
	getRec := httptest.NewRecorder()
	e.ServeHTTP(getRec, getReq)
	csrfToken := getCsrfToken(getRec)

	// Post empty
	form := url.Values{
		"content": {""},
		"_csrf":   {csrfToken},
	}
	rec := doPost(t, e, "/forum/topics/"+strconv.FormatInt(topicID, 10)+"/posts", form, sessID, csrfToken, true)
	if rec.Code != http.StatusUnprocessableEntity {
		t.Errorf("expected 422, got %d", rec.Code)
	}
}

func TestCreateSection_ModeratorSuccess(t *testing.T) {
	pool := testDB(t)
	cleanForumData(t, pool)
	defer cleanForumData(t, pool)

	_, sessID := createUser(t, pool, "forumtest-mod@test.com", "mod", "moderator")

	e := newTestServer(t, pool)

	// Get CSRF token
	getReq := httptest.NewRequest(http.MethodGet, "/forum/sections/new", nil)
	getReq.Header.Set("Cookie", "session_id="+sessID)
	getRec := httptest.NewRecorder()
	e.ServeHTTP(getRec, getReq)
	csrfToken := getCsrfToken(getRec)

	// Post
	form := url.Values{
		"title":       {"test-section"},
		"description": {"Test description"},
		"sort_order":  {"0"},
		"_csrf":       {csrfToken},
	}
	rec := doPost(t, e, "/forum/sections", form, sessID, csrfToken, false)
	if rec.Code != http.StatusSeeOther {
		t.Errorf("expected 303, got %d", rec.Code)
	}
}

func TestCreateSection_UserForbidden(t *testing.T) {
	pool := testDB(t)
	cleanForumData(t, pool)
	defer cleanForumData(t, pool)

	_, sessID := createUser(t, pool, "forumtest-user2@test.com", "user2", "user")

	e := newTestServer(t, pool)

	// Get CSRF token (this should work — RequireAuth passes, but NewSection handler redirects)
	// Actually, trying to GET /forum/sections/new as user should hit RequireRole and return 403
	getReq := httptest.NewRequest(http.MethodGet, "/forum/sections/new", nil)
	getReq.Header.Set("Cookie", "session_id="+sessID)
	getRec := httptest.NewRecorder()
	e.ServeHTTP(getRec, getReq)
	if getRec.Code != http.StatusForbidden {
		t.Errorf("expected 403 for user, got %d", getRec.Code)
	}
}

func TestIndex_HTMX(t *testing.T) {
	pool := testDB(t)
	cleanForumData(t, pool)
	defer cleanForumData(t, pool)

	e := newTestServer(t, pool)

	req := httptest.NewRequest(http.MethodGet, "/forum", nil)
	req.Header.Set("HX-Request", "true")
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

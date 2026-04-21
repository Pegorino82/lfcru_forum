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
	forumHub := forum.NewHub()
	forumHandler := forum.NewHandler(forumSvc, renderer, forumHub)

	modGroup := e.Group("", auth.RequireAuth, auth.RequireRole(renderer, "moderator", "admin"))
	modGroup.GET("/forum/sections/new", forumHandler.NewSection)
	modGroup.POST("/forum/sections", forumHandler.CreateSection)
	modGroup.GET("/forum/sections/:id/topics/new", forumHandler.NewTopic)
	modGroup.POST("/forum/sections/:id/topics", forumHandler.CreateTopic)

	e.GET("/forum", forumHandler.Index)
	e.GET("/forum/sections/:id", forumHandler.ShowSection)
	e.GET("/forum/topics/:id", forumHandler.ShowTopic)
	e.GET("/forum/topics/:id/events", forumHandler.StreamEvents)

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

// ─── SSE Tests ───────────────────────────────────────────────────────────────

// TestStreamEvents_LiveBroadcast verifies CHK-01: SSE endpoint returns text/event-stream
// and delivers a post-added event after CreatePost is called.
func TestStreamEvents_LiveBroadcast(t *testing.T) {
	pool := testDB(t)
	cleanForumData(t, pool)
	defer cleanForumData(t, pool)

	userID, sessID := createUser(t, pool, "forumtest-sse1@test.com", "sseuser1", "user")
	otherID, _ := createUser(t, pool, "forumtest-sse2@test.com", "sseuser2", "user")
	sectionID := insertTestSection(t, pool, "test-sse-section", "", 0)
	topicID := insertTestTopic(t, pool, sectionID, userID, "test-sse-topic")

	e := newTestServer(t, pool)

	// Open SSE connection as otherID (subscriber)
	sseReq := httptest.NewRequest(http.MethodGet, "/forum/topics/"+strconv.FormatInt(topicID, 10)+"/events", nil)
	sseRec := httptest.NewRecorder()

	// Run SSE handler in background; cancel after we receive data
	sseCtx, sseCancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer sseCancel()
	sseReq = sseReq.WithContext(sseCtx)

	done := make(chan struct{})
	go func() {
		defer close(done)
		e.ServeHTTP(sseRec, sseReq)
	}()

	// Give handler time to subscribe and flush headers
	time.Sleep(50 * time.Millisecond)

	if ct := sseRec.Header().Get("Content-Type"); ct != "text/event-stream" {
		t.Fatalf("expected Content-Type text/event-stream, got %q", ct)
	}

	// Post as userID (author — should NOT receive own post)
	// We just need otherID subscriber to receive it
	getReq := httptest.NewRequest(http.MethodGet, "/forum/topics/"+strconv.FormatInt(topicID, 10), nil)
	getReq.Header.Set("Cookie", "session_id="+sessID)
	getRec := httptest.NewRecorder()
	e.ServeHTTP(getRec, getReq)
	csrfToken := getCsrfToken(getRec)

	form := url.Values{"content": {"hello from sse test"}, "_csrf": {csrfToken}}
	doPost(t, e, "/forum/topics/"+strconv.FormatInt(topicID, 10)+"/posts", form, sessID, csrfToken, true)

	// Wait for SSE context to expire
	<-done

	body := sseRec.Body.String()
	if !strings.Contains(body, "event: post-added") {
		t.Errorf("expected 'event: post-added' in SSE stream, got:\n%s", body)
	}
	if !strings.Contains(body, "hello from sse test") {
		t.Errorf("expected post content in SSE stream, got:\n%s", body)
	}
	_ = otherID
}

// TestStreamEvents_AnonymousAccess verifies CHK-05: SSE endpoint returns 200 without a session cookie.
func TestStreamEvents_AnonymousAccess(t *testing.T) {
	pool := testDB(t)
	cleanForumData(t, pool)
	defer cleanForumData(t, pool)

	authorID, _ := createUser(t, pool, "forumtest-sseanon@test.com", "sseanon", "user")
	sectionID := insertTestSection(t, pool, "test-sse-anon-section", "", 0)
	topicID := insertTestTopic(t, pool, sectionID, authorID, "test-sse-anon-topic")

	e := newTestServer(t, pool)

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()
	req := httptest.NewRequest(http.MethodGet, "/forum/topics/"+strconv.FormatInt(topicID, 10)+"/events", nil)
	req = req.WithContext(ctx)
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200 for anonymous SSE, got %d", rec.Code)
	}
	if ct := rec.Header().Get("Content-Type"); ct != "text/event-stream" {
		t.Errorf("expected text/event-stream, got %q", ct)
	}
}

// TestStreamEvents_SubscriberLimit verifies CHK-06: 201st subscriber gets 503.
func TestStreamEvents_SubscriberLimit(t *testing.T) {
	pool := testDB(t)
	cleanForumData(t, pool)
	defer cleanForumData(t, pool)

	authorID, _ := createUser(t, pool, "forumtest-sselimit@test.com", "sselimit", "user")
	sectionID := insertTestSection(t, pool, "test-sse-limit-section", "", 0)
	topicID := insertTestTopic(t, pool, sectionID, authorID, "test-sse-limit-topic")

	e := newTestServer(t, pool)

	// Subscribe maxSubscribersPerTopic (200) connections
	rootCtx, rootCancel := context.WithCancel(context.Background())
	defer rootCancel()

	for i := 0; i < forum.MaxSubscribersPerTopic; i++ {
		ctx, cancel := context.WithCancel(rootCtx)
		_ = cancel
		req := httptest.NewRequest(http.MethodGet, "/forum/topics/"+strconv.FormatInt(topicID, 10)+"/events", nil)
		req = req.WithContext(ctx)
		rec := httptest.NewRecorder()
		go e.ServeHTTP(rec, req)
	}

	// Give goroutines time to subscribe
	time.Sleep(100 * time.Millisecond)

	// 201st request must return 503
	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()
	req := httptest.NewRequest(http.MethodGet, "/forum/topics/"+strconv.FormatInt(topicID, 10)+"/events", nil)
	req = req.WithContext(ctx)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusServiceUnavailable {
		t.Errorf("expected 503 for 201st subscriber, got %d", rec.Code)
	}
}

// TestStreamEvents_CatchUp verifies CHK-02: posts published during disconnect are delivered via Last-Event-ID.
func TestStreamEvents_CatchUp(t *testing.T) {
	pool := testDB(t)
	cleanForumData(t, pool)
	defer cleanForumData(t, pool)

	userID, sessID := createUser(t, pool, "forumtest-ssecatchup@test.com", "ssecatchup", "user")
	sectionID := insertTestSection(t, pool, "test-sse-catchup-section", "", 0)
	topicID := insertTestTopic(t, pool, sectionID, userID, "test-sse-catchup-topic")

	e := newTestServer(t, pool)

	// Publish a post via HTTP (simulates missed post during disconnect)
	getReq := httptest.NewRequest(http.MethodGet, "/forum/topics/"+strconv.FormatInt(topicID, 10), nil)
	getReq.Header.Set("Cookie", "session_id="+sessID)
	getRec := httptest.NewRecorder()
	e.ServeHTTP(getRec, getReq)
	csrfToken := getCsrfToken(getRec)

	form := url.Values{"content": {"missed post content"}, "_csrf": {csrfToken}}
	postRec := doPost(t, e, "/forum/topics/"+strconv.FormatInt(topicID, 10)+"/posts", form, sessID, csrfToken, true)
	if postRec.Code != http.StatusCreated {
		t.Fatalf("create post: expected 201, got %d", postRec.Code)
	}

	// Reconnect with Last-Event-ID=0 to catch up all posts
	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()
	req := httptest.NewRequest(http.MethodGet, "/forum/topics/"+strconv.FormatInt(topicID, 10)+"/events", nil)
	req.Header.Set("Last-Event-ID", "0")
	req = req.WithContext(ctx)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	body := rec.Body.String()
	if !strings.Contains(body, "event: post-added") {
		t.Errorf("expected 'event: post-added' in catch-up response, got:\n%s", body)
	}
	if !strings.Contains(body, "missed post content") {
		t.Errorf("expected missed post content in catch-up response, got:\n%s", body)
	}
}

// Regression: FT-013 — forum pages must display logged-in user in navigation.
// Before fix: forum handler did not pass User to data map → nav showed «Войти / Регистрация»
// for authenticated users on all forum pages.
func TestIndex_AuthUser_ShowsUsername(t *testing.T) {
	pool := testDB(t)
	cleanForumData(t, pool)
	defer cleanForumData(t, pool)

	_, sessID := createUser(t, pool, "forumtest-navuser@test.com", "navuser", "user")

	e := newTestServer(t, pool)
	rec := doGet(t, e, "/forum", sessID)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
	body := rec.Body.String()
	if !strings.Contains(body, "navuser") {
		t.Error("expected username in navigation for authenticated user (FT-013 regression)")
	}
}

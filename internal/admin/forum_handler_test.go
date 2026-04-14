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
	"github.com/Pegorino82/lfcru_forum/internal/forum"
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

// newForumAdminServer builds an Echo server with admin forum routes registered.
func newForumAdminServer(t *testing.T, pool *pgxpool.Pool) *echo.Echo {
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

	forumRepo := forum.NewRepo(pool)
	forumSvc := forum.NewService(forumRepo)
	forumAdminHandler := admin.NewForumHandler(forumSvc)

	e := echo.New()
	e.HideBanner = true
	e.Renderer = renderer
	e.Use(middleware.Recover())
	e.Use(appMiddleware.CSRFMiddleware())
	e.Use(auth.LoadSession(authSvc))

	adminGroup := e.Group("", admin.RequireAdminOrMod(renderer))
	adminGroup.GET("/admin/forum/sections", forumAdminHandler.ListSections)
	adminGroup.GET("/admin/forum/sections/new", forumAdminHandler.NewSection)
	adminGroup.POST("/admin/forum/sections", forumAdminHandler.CreateSection)
	adminGroup.GET("/admin/forum/sections/:id/edit", forumAdminHandler.EditSection)
	adminGroup.POST("/admin/forum/sections/:id", forumAdminHandler.UpdateSection)
	adminGroup.GET("/admin/forum/sections/:id/topics", forumAdminHandler.ListTopics)
	adminGroup.GET("/admin/forum/sections/:id/topics/new", forumAdminHandler.NewTopic)
	adminGroup.POST("/admin/forum/sections/:id/topics", forumAdminHandler.CreateTopic)
	adminGroup.GET("/admin/forum/topics/:id/edit", forumAdminHandler.EditTopic)
	adminGroup.POST("/admin/forum/topics/:id", forumAdminHandler.UpdateTopic)

	return e
}

func cleanForumAdminData(t *testing.T, pool *pgxpool.Pool) {
	t.Helper()
	ctx := context.Background()
	pool.Exec(ctx, `DELETE FROM forum_topics WHERE title LIKE 'admtest-%'`)
	pool.Exec(ctx, `DELETE FROM forum_sections WHERE title LIKE 'admtest-%'`)
	pool.Exec(ctx, `DELETE FROM sessions WHERE user_id IN (SELECT id FROM users WHERE email LIKE 'fadmintest%')`)
	pool.Exec(ctx, `DELETE FROM users WHERE email LIKE 'fadmintest%'`)
}

func insertAdminSection(t *testing.T, pool *pgxpool.Pool, title, description string) int64 {
	t.Helper()
	var id int64
	if err := pool.QueryRow(context.Background(),
		`INSERT INTO forum_sections (title, description, sort_order) VALUES ($1, $2, 0) RETURNING id`,
		title, description,
	).Scan(&id); err != nil {
		t.Fatalf("insert section: %v", err)
	}
	return id
}

func insertAdminTopic(t *testing.T, pool *pgxpool.Pool, sectionID, authorID int64, title string) int64 {
	t.Helper()
	var id int64
	if err := pool.QueryRow(context.Background(),
		`INSERT INTO forum_topics (section_id, author_id, title) VALUES ($1, $2, $3) RETURNING id`,
		sectionID, authorID, title,
	).Scan(&id); err != nil {
		t.Fatalf("insert topic: %v", err)
	}
	return id
}

// getForumCsrf fetches the CSRF token by making a GET to the sections list.
func getForumCsrf(t *testing.T, e *echo.Echo, sessID string) string {
	t.Helper()
	rec := doGet(t, e, "/admin/forum/sections", sessID)
	for _, c := range rec.Result().Cookies() {
		if c.Name == "_csrf" {
			return c.Value
		}
	}
	t.Fatal("no _csrf cookie in response")
	return ""
}

// doForumPost performs a POST with session + CSRF cookies and form body.
func doForumPost(t *testing.T, e *echo.Echo, path string, form url.Values, sessID, csrfToken string) *httptest.ResponseRecorder {
	t.Helper()
	form.Set("_csrf", csrfToken)
	req := httptest.NewRequest(http.MethodPost, path, strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Cookie", "session_id="+sessID+"; _csrf="+csrfToken)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	return rec
}

// ─── Tests ───────────────────────────────────────────────────────────────────

// SC-01: Admin открывает /admin/forum/sections → 200 со списком разделов.
func TestAdminForum_ListSections(t *testing.T) {
	pool := testDB(t)
	cleanForumAdminData(t, pool)
	e := newForumAdminServer(t, pool)

	_, sessID := createUserWithRole(t, pool, "fadmintest-a1@test.com", "fadmintest_a1", "admin")
	insertAdminSection(t, pool, "admtest-раздел1", "описание")

	rec := doGet(t, e, "/admin/forum/sections", sessID)
	if rec.Code != http.StatusOK {
		t.Fatalf("want 200, got %d\n%s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), "admtest-раздел1") {
		t.Fatal("response does not contain section title")
	}
}

// SC-02: Admin создаёт раздел → он появляется в БД.
func TestAdminForum_CreateSection(t *testing.T) {
	pool := testDB(t)
	cleanForumAdminData(t, pool)
	e := newForumAdminServer(t, pool)

	_, sessID := createUserWithRole(t, pool, "fadmintest-a2@test.com", "fadmintest_a2", "admin")
	csrfToken := getForumCsrf(t, e, sessID)

	form := url.Values{
		"title":       {"admtest-новый раздел"},
		"description": {"описание нового раздела"},
	}
	rec := doForumPost(t, e, "/admin/forum/sections", form, sessID, csrfToken)
	if rec.Code != http.StatusSeeOther {
		t.Fatalf("want 303, got %d\n%s", rec.Code, rec.Body.String())
	}

	var cnt int
	pool.QueryRow(context.Background(),
		`SELECT COUNT(*) FROM forum_sections WHERE title = 'admtest-новый раздел'`,
	).Scan(&cnt)
	if cnt == 0 {
		t.Fatal("section not created in DB")
	}
}

// EC-04: Попытка создать раздел с пустым именем → 400.
func TestAdminForum_CreateSection_EmptyTitle(t *testing.T) {
	pool := testDB(t)
	cleanForumAdminData(t, pool)
	e := newForumAdminServer(t, pool)

	_, sessID := createUserWithRole(t, pool, "fadmintest-a3@test.com", "fadmintest_a3", "admin")
	csrfToken := getForumCsrf(t, e, sessID)

	rec := doForumPost(t, e, "/admin/forum/sections", url.Values{"title": {""}}, sessID, csrfToken)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("want 400, got %d", rec.Code)
	}
}

// SC-03: Admin редактирует название раздела → изменение отражается в БД.
func TestAdminForum_UpdateSection(t *testing.T) {
	pool := testDB(t)
	cleanForumAdminData(t, pool)
	e := newForumAdminServer(t, pool)

	_, sessID := createUserWithRole(t, pool, "fadmintest-a4@test.com", "fadmintest_a4", "admin")
	sectionID := insertAdminSection(t, pool, "admtest-старое название", "")
	csrfToken := getForumCsrf(t, e, sessID)

	form := url.Values{
		"title":       {"admtest-новое название"},
		"description": {""},
	}
	rec := doForumPost(t, e, "/admin/forum/sections/"+strconv.FormatInt(sectionID, 10), form, sessID, csrfToken)
	if rec.Code != http.StatusSeeOther {
		t.Fatalf("want 303, got %d\n%s", rec.Code, rec.Body.String())
	}

	var title string
	pool.QueryRow(context.Background(),
		`SELECT title FROM forum_sections WHERE id = $1`, sectionID,
	).Scan(&title)
	if title != "admtest-новое название" {
		t.Fatalf("want updated title, got %q", title)
	}
}

// SC-04: Admin открывает список тем раздела → 200 с темами.
func TestAdminForum_ListTopics(t *testing.T) {
	pool := testDB(t)
	cleanForumAdminData(t, pool)
	e := newForumAdminServer(t, pool)

	adminID, sessID := createUserWithRole(t, pool, "fadmintest-a5@test.com", "fadmintest_a5", "admin")
	sectionID := insertAdminSection(t, pool, "admtest-раздел-topics", "")
	insertAdminTopic(t, pool, sectionID, adminID, "admtest-тема1")

	rec := doGet(t, e, "/admin/forum/sections/"+strconv.FormatInt(sectionID, 10)+"/topics", sessID)
	if rec.Code != http.StatusOK {
		t.Fatalf("want 200, got %d\n%s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), "admtest-тема1") {
		t.Fatal("response does not contain topic title")
	}
}

// SC-05: Admin создаёт тему в разделе → тема появляется в БД.
func TestAdminForum_CreateTopic(t *testing.T) {
	pool := testDB(t)
	cleanForumAdminData(t, pool)
	e := newForumAdminServer(t, pool)

	_, sessID := createUserWithRole(t, pool, "fadmintest-a6@test.com", "fadmintest_a6", "admin")
	sectionID := insertAdminSection(t, pool, "admtest-раздел-create-topic", "")
	csrfToken := getForumCsrf(t, e, sessID)

	form := url.Values{"title": {"admtest-новая тема"}}
	rec := doForumPost(t, e, "/admin/forum/sections/"+strconv.FormatInt(sectionID, 10)+"/topics", form, sessID, csrfToken)
	if rec.Code != http.StatusSeeOther {
		t.Fatalf("want 303, got %d\n%s", rec.Code, rec.Body.String())
	}

	var cnt int
	pool.QueryRow(context.Background(),
		`SELECT COUNT(*) FROM forum_topics WHERE title = 'admtest-новая тема' AND section_id = $1`, sectionID,
	).Scan(&cnt)
	if cnt == 0 {
		t.Fatal("topic not created in DB")
	}
}

// SC-06: Admin редактирует название темы → изменение отражается в БД.
func TestAdminForum_UpdateTopic(t *testing.T) {
	pool := testDB(t)
	cleanForumAdminData(t, pool)
	e := newForumAdminServer(t, pool)

	adminID, sessID := createUserWithRole(t, pool, "fadmintest-a7@test.com", "fadmintest_a7", "admin")
	sectionID := insertAdminSection(t, pool, "admtest-раздел-update-topic", "")
	topicID := insertAdminTopic(t, pool, sectionID, adminID, "admtest-старая тема")
	csrfToken := getForumCsrf(t, e, sessID)

	form := url.Values{"title": {"admtest-обновлённая тема"}}
	rec := doForumPost(t, e, "/admin/forum/topics/"+strconv.FormatInt(topicID, 10), form, sessID, csrfToken)
	if rec.Code != http.StatusSeeOther {
		t.Fatalf("want 303, got %d\n%s", rec.Code, rec.Body.String())
	}

	var title string
	pool.QueryRow(context.Background(),
		`SELECT title FROM forum_topics WHERE id = $1`, topicID,
	).Scan(&title)
	if title != "admtest-обновлённая тема" {
		t.Fatalf("want updated title, got %q", title)
	}
}

// FM-02: Раздел не найден → 404.
func TestAdminForum_ListTopics_SectionNotFound(t *testing.T) {
	pool := testDB(t)
	cleanForumAdminData(t, pool)
	e := newForumAdminServer(t, pool)

	_, sessID := createUserWithRole(t, pool, "fadmintest-a8@test.com", "fadmintest_a8", "admin")

	rec := doGet(t, e, "/admin/forum/sections/999999/topics", sessID)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("want 404, got %d", rec.Code)
	}
}

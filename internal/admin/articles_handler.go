package admin

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/Pegorino82/lfcru_forum/internal/auth"
	appMiddleware "github.com/Pegorino82/lfcru_forum/internal/middleware"
	"github.com/Pegorino82/lfcru_forum/internal/news"
	"github.com/Pegorino82/lfcru_forum/internal/user"
	"github.com/labstack/echo/v4"
)

// ArticlesRepo is the subset of news.Repo the admin articles handler needs.
type ArticlesRepo interface {
	CreateDraft(ctx context.Context, n *news.News) error
	UpdateArticle(ctx context.Context, n *news.News) error
	ChangeStatus(ctx context.Context, id int64, status news.ArticleStatus, reviewerID *int64) error
	ListByStatus(ctx context.Context, status news.ArticleStatus) ([]news.News, error)
	GetByIDAdmin(ctx context.Context, id int64) (*news.News, error)
	ListImagesByArticleID(ctx context.Context, articleID int64) ([]news.ImageView, error)
}

// ArticlesHandler handles admin article management routes.
type ArticlesHandler struct {
	repo       ArticlesRepo
	imagesRepo *ImagesRepo
}

// NewArticlesHandler creates a new ArticlesHandler.
func NewArticlesHandler(repo ArticlesRepo, imagesRepo *ImagesRepo) *ArticlesHandler {
	return &ArticlesHandler{repo: repo, imagesRepo: imagesRepo}
}

// validTransitions defines the allowed status transitions.
var validTransitions = map[news.ArticleStatus]map[news.ArticleStatus]bool{
	news.StatusDraft:     {news.StatusInReview: true, news.StatusPublished: true},
	news.StatusInReview:  {news.StatusPublished: true},
	news.StatusPublished: {news.StatusDraft: true},
}

type articlesListData struct {
	User      *user.User
	CSRFToken string
	Articles  []news.News
	Filter    string
}

type articleEditData struct {
	User      *user.User
	CSRFToken string
	Article   *news.News
	Images    []ArticleImage
	Error     string
	Saved     bool
}

type articlePreviewData struct {
	User        *user.User
	CSRFToken   string
	Article     *news.News
	ContentHTML interface{}
	Comments    []struct{}
	Images      []news.ImageView
	NewsID      int64
	IsPreview   bool
}

// List handles GET /admin/articles.
func (h *ArticlesHandler) List(c echo.Context) error {
	filter := news.ArticleStatus(c.QueryParam("status"))
	articles, err := h.repo.ListByStatus(c.Request().Context(), filter)
	if err != nil {
		slog.Error("admin: list articles", "error", err)
		return c.String(http.StatusInternalServerError, "Внутренняя ошибка сервера")
	}
	return c.Render(http.StatusOK, "templates/admin/articles/list.html", articlesListData{
		User:      auth.UserFromContext(c),
		CSRFToken: appMiddleware.CSRFToken(c),
		Articles:  articles,
		Filter:    string(filter),
	})
}

// New handles GET /admin/articles/new.
func (h *ArticlesHandler) New(c echo.Context) error {
	return c.Render(http.StatusOK, "templates/admin/articles/edit.html", articleEditData{
		User:      auth.UserFromContext(c),
		CSRFToken: appMiddleware.CSRFToken(c),
	})
}

// Create handles POST /admin/articles.
func (h *ArticlesHandler) Create(c echo.Context) error {
	cu := auth.UserFromContext(c)
	title := c.FormValue("title")
	content := c.FormValue("content")

	if title == "" {
		return c.Render(http.StatusUnprocessableEntity, "templates/admin/articles/edit.html", articleEditData{
			User:      cu,
			CSRFToken: appMiddleware.CSRFToken(c),
			Error:     "Заголовок обязателен",
		})
	}

	n := &news.News{Title: title, Content: content, AuthorID: cu.ID}
	if err := h.repo.CreateDraft(c.Request().Context(), n); err != nil {
		slog.Error("admin: create article", "error", err)
		return c.String(http.StatusInternalServerError, "Внутренняя ошибка сервера")
	}
	return c.Redirect(http.StatusSeeOther, "/admin/articles/"+strconv.FormatInt(n.ID, 10)+"/edit")
}

// Edit handles GET /admin/articles/:id/edit.
func (h *ArticlesHandler) Edit(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		return c.String(http.StatusBadRequest, "Некорректный ID")
	}
	ctx := c.Request().Context()

	article, err := h.repo.GetByIDAdmin(ctx, id)
	if err != nil {
		slog.Error("admin: get article", "error", err)
		return c.String(http.StatusInternalServerError, "Внутренняя ошибка сервера")
	}
	if article == nil {
		return c.String(http.StatusNotFound, "Статья не найдена")
	}

	images, err := h.imagesRepo.ListByArticleID(ctx, id)
	if err != nil {
		slog.Error("admin: list images", "error", err)
		return c.String(http.StatusInternalServerError, "Внутренняя ошибка сервера")
	}

	return c.Render(http.StatusOK, "templates/admin/articles/edit.html", articleEditData{
		User:      auth.UserFromContext(c),
		CSRFToken: appMiddleware.CSRFToken(c),
		Article:   article,
		Images:    images,
		Saved:     c.QueryParam("saved") == "1",
	})
}

// Update handles POST /admin/articles/:id.
func (h *ArticlesHandler) Update(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		return c.String(http.StatusBadRequest, "Некорректный ID")
	}
	ctx := c.Request().Context()

	title := c.FormValue("title")
	content := c.FormValue("content")

	if title == "" {
		article, _ := h.repo.GetByIDAdmin(ctx, id)
		images, _ := h.imagesRepo.ListByArticleID(ctx, id)
		return c.Render(http.StatusUnprocessableEntity, "templates/admin/articles/edit.html", articleEditData{
			User:      auth.UserFromContext(c),
			CSRFToken: appMiddleware.CSRFToken(c),
			Article:   article,
			Images:    images,
			Error:     "Заголовок обязателен",
		})
	}

	n := &news.News{ID: id, Title: title, Content: content}
	if err := h.repo.UpdateArticle(ctx, n); err != nil {
		slog.Error("admin: update article", "error", err)
		return c.String(http.StatusInternalServerError, "Внутренняя ошибка сервера")
	}
	return c.Redirect(http.StatusSeeOther, "/admin/articles/"+strconv.FormatInt(id, 10)+"/edit?saved=1")
}

// Preview handles GET /admin/articles/:id/preview.
func (h *ArticlesHandler) Preview(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		return c.String(http.StatusBadRequest, "Некорректный ID")
	}
	ctx := c.Request().Context()

	article, err := h.repo.GetByIDAdmin(ctx, id)
	if err != nil {
		slog.Error("admin: preview article", "error", err)
		return c.String(http.StatusInternalServerError, "Внутренняя ошибка сервера")
	}
	if article == nil {
		return c.String(http.StatusNotFound, "Статья не найдена")
	}

	images, err := h.repo.ListImagesByArticleID(ctx, id)
	if err != nil {
		slog.Error("admin: preview images", "error", err)
		return c.String(http.StatusInternalServerError, "Внутренняя ошибка сервера")
	}

	data := articlePreviewData{
		User:        auth.UserFromContext(c),
		CSRFToken:   appMiddleware.CSRFToken(c),
		Article:     article,
		ContentHTML: news.RenderMarkdown(article.Content),
		Images:      images,
		NewsID:      id,
		IsPreview:   true,
	}
	return c.Render(http.StatusOK, "templates/news/article.html", data)
}

// ChangeStatus handles POST /admin/articles/:id/status.
func (h *ArticlesHandler) ChangeStatus(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		return c.String(http.StatusBadRequest, "Некорректный ID")
	}
	ctx := c.Request().Context()

	newStatus := news.ArticleStatus(c.FormValue("status"))

	article, err := h.repo.GetByIDAdmin(ctx, id)
	if err != nil {
		slog.Error("admin: change status get", "error", err)
		return c.String(http.StatusInternalServerError, "Внутренняя ошибка сервера")
	}
	if article == nil {
		return c.String(http.StatusNotFound, "Статья не найдена")
	}

	if !validTransitions[article.Status][newStatus] {
		return c.String(http.StatusBadRequest, "Недопустимый переход статуса")
	}

	// Для публикации нужен непустой заголовок (FM-01).
	if newStatus == news.StatusPublished && (article.Title == "" || article.Content == "") {
		return c.String(http.StatusBadRequest, "Невозможно опубликовать статью без заголовка или текста")
	}

	var reviewerID *int64
	if newStatus == news.StatusInReview {
		if revStr := c.FormValue("reviewer_id"); revStr != "" {
			revID, parseErr := strconv.ParseInt(revStr, 10, 64)
			if parseErr == nil && revID > 0 {
				reviewerID = &revID
			}
		}
	}

	if err := h.repo.ChangeStatus(ctx, id, newStatus, reviewerID); err != nil {
		if errors.Is(err, errors.New("no rows")) {
			return c.String(http.StatusNotFound, "Статья не найдена")
		}
		slog.Error("admin: change status", "error", err)
		return c.String(http.StatusInternalServerError, "Внутренняя ошибка сервера")
	}

	return c.Redirect(http.StatusSeeOther, "/admin/articles/"+strconv.FormatInt(id, 10)+"/edit")
}

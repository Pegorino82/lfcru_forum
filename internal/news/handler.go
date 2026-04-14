package news

import (
	"context"
	"errors"
	"fmt"
	"html/template"
	"log/slog"
	"net/http"
	"net/url"
	"strconv"

	"github.com/Pegorino82/lfcru_forum/internal/auth"
	"github.com/Pegorino82/lfcru_forum/internal/comment"
	appMiddleware "github.com/Pegorino82/lfcru_forum/internal/middleware"
	"github.com/Pegorino82/lfcru_forum/internal/tmpl"
	"github.com/Pegorino82/lfcru_forum/internal/user"
	"github.com/labstack/echo/v4"
)

const pageSize = 20

// ListData is the template data for the news list page.
type ListData struct {
	User       *user.User
	Items      []News
	Page       int
	TotalPages int
	HasPrev    bool
	HasNext    bool
}

// ArticleData is the template data for the article page.
type ArticleData struct {
	User        *user.User
	CSRFToken   string
	Article     *News
	ContentHTML template.HTML
	Comments    []comment.CommentView
	Images      []ImageView
	NewsID      int64
}

// Handler handles news HTTP requests.
type Handler struct {
	newsRepo    *Repo
	commentRepo *comment.Repo
	commentSvc  *comment.Service
}

// NewHandler creates a new news Handler.
func NewHandler(newsRepo *Repo, commentRepo *comment.Repo, commentSvc *comment.Service) *Handler {
	return &Handler{newsRepo: newsRepo, commentRepo: commentRepo, commentSvc: commentSvc}
}

// RegisterRoutes mounts news routes on the Echo instance.
func (h *Handler) RegisterRoutes(e *echo.Echo) {
	e.GET("/news", h.ShowList)
	e.GET("/news/:id", h.ShowArticle)
	e.POST("/news/:id/comments", h.CreateComment)
}

// ShowList renders the paginated news list.
func (h *Handler) ShowList(c echo.Context) error {
	ctx := c.Request().Context()

	page, _ := strconv.Atoi(c.QueryParam("page"))
	if page < 1 {
		page = 1
	}
	offset := (page - 1) * pageSize

	items, total, err := h.newsRepo.ListPublished(ctx, pageSize, offset)
	if err != nil {
		slog.Error("news list: load", "err", err)
		return c.String(http.StatusInternalServerError, "Что-то пошло не так. Попробуйте обновить страницу.")
	}

	totalPages := int((total + int64(pageSize) - 1) / int64(pageSize))
	if totalPages < 1 {
		totalPages = 1
	}

	data := ListData{
		User:       auth.UserFromContext(c),
		Items:      items,
		Page:       page,
		TotalPages: totalPages,
		HasPrev:    page > 1,
		HasNext:    page < totalPages,
	}

	if c.Request().Header.Get("HX-Request") == "true" {
		r := c.Echo().Renderer.(*tmpl.Renderer)
		return r.RenderPartial(c.Response(), "templates/news/list.html", "content", data)
	}
	return c.Render(http.StatusOK, "templates/news/list.html", data)
}

// ShowArticle renders the article page with comments.
func (h *Handler) ShowArticle(c echo.Context) error {
	ctx := c.Request().Context()

	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		return c.String(http.StatusNotFound, "Статья не найдена")
	}

	article, err := h.newsRepo.GetPublishedByID(ctx, id)
	if err != nil {
		slog.Error("article: load", "id", id, "err", err)
		return c.String(http.StatusInternalServerError, "Что-то пошло не так. Попробуйте обновить страницу.")
	}
	if article == nil {
		return c.String(http.StatusNotFound, "Статья не найдена")
	}

	comments, err := h.commentRepo.ListByNewsID(ctx, id)
	if err != nil {
		slog.Error("article: load comments", "id", id, "err", err)
		return c.String(http.StatusInternalServerError, "Что-то пошло не так. Попробуйте обновить страницу.")
	}

	if err := h.fillMentions(ctx, comments); err != nil {
		slog.Error("article: render mentions", "err", err)
		return c.String(http.StatusInternalServerError, "Что-то пошло не так. Попробуйте обновить страницу.")
	}

	images, err := h.newsRepo.ListImagesByArticleID(ctx, id)
	if err != nil {
		slog.Error("article: load images", "id", id, "err", err)
		return c.String(http.StatusInternalServerError, "Что-то пошло не так. Попробуйте обновить страницу.")
	}

	data := ArticleData{
		User:        auth.UserFromContext(c),
		CSRFToken:   appMiddleware.CSRFToken(c),
		Article:     article,
		ContentHTML: RenderMarkdown(article.Content),
		Comments:    comments,
		Images:      images,
		NewsID:      id,
	}

	if c.Request().Header.Get("HX-Request") == "true" {
		r := c.Echo().Renderer.(*tmpl.Renderer)
		return r.RenderPartial(c.Response(), "templates/news/article.html", "content", data)
	}
	return c.Render(http.StatusOK, "templates/news/article.html", data)
}

// CreateComment handles comment form submission.
func (h *Handler) CreateComment(c echo.Context) error {
	ctx := c.Request().Context()
	currentUser := auth.UserFromContext(c)

	if currentUser == nil {
		next := url.QueryEscape(c.Request().URL.RequestURI())
		if c.Request().Header.Get("HX-Request") == "true" {
			c.Response().Header().Set("HX-Redirect", "/login?next="+next)
			return c.NoContent(http.StatusUnauthorized)
		}
		return c.Redirect(http.StatusFound, "/login?next="+next)
	}

	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		return c.Redirect(http.StatusSeeOther, "/")
	}

	cm := &comment.Comment{
		NewsID:   id,
		AuthorID: currentUser.ID,
		Content:  c.FormValue("content"),
	}

	if parentStr := c.FormValue("parent_id"); parentStr != "" {
		pid, pErr := strconv.ParseInt(parentStr, 10, 64)
		if pErr == nil && pid > 0 {
			cm.ParentID = &pid
		}
	}

	_, err = h.commentSvc.Create(ctx, cm)
	if err != nil {
		errMsg := mapCommentError(err)
		if errMsg == "" {
			slog.Error("create comment", "err", err)
			return c.String(http.StatusInternalServerError, "Что-то пошло не так.")
		}
		if c.Request().Header.Get("HX-Request") == "true" {
			c.Response().Header().Set("HX-Retarget", "#comment-error")
			c.Response().Header().Set("HX-Reswap", "outerHTML")
			return c.HTML(http.StatusUnprocessableEntity,
				`<p id="comment-error" class="error">`+errMsg+`</p>`)
		}
		articleFlash(c, errMsg)
		return c.Redirect(http.StatusSeeOther, fmt.Sprintf("/news/%d", id))
	}

	// Success: for HTMX requests return updated comments list.
	if c.Request().Header.Get("HX-Request") == "true" {
		return h.renderCommentsList(c, ctx, id, currentUser)
	}
	return c.Redirect(http.StatusSeeOther, fmt.Sprintf("/news/%d", id))
}

// renderCommentsList re-fetches and renders the comments-list partial for HTMX swaps.
func (h *Handler) renderCommentsList(c echo.Context, ctx context.Context, newsID int64, currentUser *user.User) error {
	comments, err := h.commentRepo.ListByNewsID(ctx, newsID)
	if err != nil {
		slog.Error("render comments list", "err", err)
		return c.String(http.StatusInternalServerError, "Что-то пошло не так.")
	}
	if err := h.fillMentions(ctx, comments); err != nil {
		slog.Error("render comments list: mentions", "err", err)
		return c.String(http.StatusInternalServerError, "Что-то пошло не так.")
	}
	data := ArticleData{
		User:      currentUser,
		CSRFToken: appMiddleware.CSRFToken(c),
		Comments:  comments,
		NewsID:    newsID,
	}
	r := c.Echo().Renderer.(*tmpl.Renderer)
	return r.RenderPartial(c.Response(), "templates/news/article.html", "comments-list", data)
}

// fillMentions calls RenderMentions for each comment and sets ContentHTML.
func (h *Handler) fillMentions(ctx context.Context, comments []comment.CommentView) error {
	for i := range comments {
		rendered, err := h.commentSvc.RenderMentions(ctx, comments[i].Content)
		if err != nil {
			return err
		}
		comments[i].ContentHTML = rendered
	}
	return nil
}

func mapCommentError(err error) string {
	switch {
	case errors.Is(err, comment.ErrEmptyContent):
		return "Комментарий не может быть пустым"
	case errors.Is(err, comment.ErrContentTooLong):
		return "Комментарий слишком длинный (максимум 10 000 символов)"
	case errors.Is(err, comment.ErrParentNotFound):
		return "Комментарий, на который вы отвечаете, не найден"
	}
	return ""
}

func articleFlash(c echo.Context, msg string) {
	c.SetCookie(&http.Cookie{
		Name:     "flash",
		Value:    url.QueryEscape(msg),
		Path:     "/",
		MaxAge:   60,
		SameSite: http.SameSiteLaxMode,
	})
}

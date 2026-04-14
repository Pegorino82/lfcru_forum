package admin

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/Pegorino82/lfcru_forum/internal/auth"
	"github.com/Pegorino82/lfcru_forum/internal/forum"
	appMiddleware "github.com/Pegorino82/lfcru_forum/internal/middleware"
	"github.com/Pegorino82/lfcru_forum/internal/user"
	"github.com/labstack/echo/v4"
)

// ForumSvc is the subset of forum.Service the admin forum handler needs.
type ForumSvc interface {
	ListSections(ctx context.Context) ([]forum.SectionView, error)
	GetSection(ctx context.Context, id int64) (*forum.Section, error)
	CreateSection(ctx context.Context, title, description string, sortOrder int) (int64, error)
	UpdateSection(ctx context.Context, id int64, title, description string) error
	ListTopicsBySection(ctx context.Context, sectionID int64) ([]forum.TopicView, error)
	GetTopic(ctx context.Context, id int64) (*forum.Topic, error)
	CreateTopic(ctx context.Context, sectionID, authorID int64, title string) (int64, error)
	UpdateTopic(ctx context.Context, id int64, title string) error
}

// ForumHandler handles admin forum management routes.
type ForumHandler struct {
	svc ForumSvc
}

// NewForumHandler creates a new ForumHandler.
func NewForumHandler(svc ForumSvc) *ForumHandler {
	return &ForumHandler{svc: svc}
}

type sectionsListData struct {
	User      *user.User
	CSRFToken string
	Sections  []forum.SectionView
}

type sectionEditData struct {
	User      *user.User
	CSRFToken string
	Section   *forum.Section
	Error     string
}

type topicsListData struct {
	User      *user.User
	CSRFToken string
	Section   *forum.Section
	Topics    []forum.TopicView
}

type topicEditData struct {
	User      *user.User
	CSRFToken string
	Section   *forum.Section
	Topic     *forum.Topic
	Error     string
}

// ListSections handles GET /admin/forum/sections.
func (h *ForumHandler) ListSections(c echo.Context) error {
	sections, err := h.svc.ListSections(c.Request().Context())
	if err != nil {
		slog.Error("admin: list sections", "error", err)
		return c.String(http.StatusInternalServerError, "Внутренняя ошибка сервера")
	}
	return c.Render(http.StatusOK, "templates/admin/forum/sections_list.html", sectionsListData{
		User:      auth.UserFromContext(c),
		CSRFToken: appMiddleware.CSRFToken(c),
		Sections:  sections,
	})
}

// NewSection handles GET /admin/forum/sections/new.
func (h *ForumHandler) NewSection(c echo.Context) error {
	return c.Render(http.StatusOK, "templates/admin/forum/section_edit.html", sectionEditData{
		User:      auth.UserFromContext(c),
		CSRFToken: appMiddleware.CSRFToken(c),
	})
}

// CreateSection handles POST /admin/forum/sections.
func (h *ForumHandler) CreateSection(c echo.Context) error {
	title := c.FormValue("title")
	description := c.FormValue("description")
	_, err := h.svc.CreateSection(c.Request().Context(), title, description, 0)
	if err != nil {
		if errors.Is(err, forum.ErrEmptyTitle) || errors.Is(err, forum.ErrTitleTooLong) || errors.Is(err, forum.ErrDescriptionTooLong) {
			return c.Render(http.StatusBadRequest, "templates/admin/forum/section_edit.html", sectionEditData{
				User:      auth.UserFromContext(c),
				CSRFToken: appMiddleware.CSRFToken(c),
				Error:     err.Error(),
			})
		}
		slog.Error("admin: create section", "error", err)
		return c.String(http.StatusInternalServerError, "Внутренняя ошибка сервера")
	}
	return c.Redirect(http.StatusSeeOther, "/admin/forum/sections")
}

// EditSection handles GET /admin/forum/sections/:id/edit.
func (h *ForumHandler) EditSection(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return c.String(http.StatusBadRequest, "Неверный ID")
	}
	section, err := h.svc.GetSection(c.Request().Context(), id)
	if err != nil {
		slog.Error("admin: edit section get", "error", err)
		return c.String(http.StatusInternalServerError, "Внутренняя ошибка сервера")
	}
	if section == nil {
		return c.String(http.StatusNotFound, "Раздел не найден")
	}
	return c.Render(http.StatusOK, "templates/admin/forum/section_edit.html", sectionEditData{
		User:      auth.UserFromContext(c),
		CSRFToken: appMiddleware.CSRFToken(c),
		Section:   section,
	})
}

// UpdateSection handles POST /admin/forum/sections/:id.
func (h *ForumHandler) UpdateSection(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return c.String(http.StatusBadRequest, "Неверный ID")
	}
	section, err := h.svc.GetSection(c.Request().Context(), id)
	if err != nil {
		slog.Error("admin: update section get", "error", err)
		return c.String(http.StatusInternalServerError, "Внутренняя ошибка сервера")
	}
	if section == nil {
		return c.String(http.StatusNotFound, "Раздел не найден")
	}

	title := c.FormValue("title")
	description := c.FormValue("description")
	if err := h.svc.UpdateSection(c.Request().Context(), id, title, description); err != nil {
		if errors.Is(err, forum.ErrEmptyTitle) || errors.Is(err, forum.ErrTitleTooLong) || errors.Is(err, forum.ErrDescriptionTooLong) {
			return c.Render(http.StatusBadRequest, "templates/admin/forum/section_edit.html", sectionEditData{
				User:      auth.UserFromContext(c),
				CSRFToken: appMiddleware.CSRFToken(c),
				Section:   section,
				Error:     err.Error(),
			})
		}
		slog.Error("admin: update section", "error", err)
		return c.String(http.StatusInternalServerError, "Внутренняя ошибка сервера")
	}
	return c.Redirect(http.StatusSeeOther, "/admin/forum/sections")
}

// ListTopics handles GET /admin/forum/sections/:id/topics.
func (h *ForumHandler) ListTopics(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return c.String(http.StatusBadRequest, "Неверный ID")
	}
	section, err := h.svc.GetSection(c.Request().Context(), id)
	if err != nil {
		slog.Error("admin: list topics get section", "error", err)
		return c.String(http.StatusInternalServerError, "Внутренняя ошибка сервера")
	}
	if section == nil {
		return c.String(http.StatusNotFound, "Раздел не найден")
	}
	topics, err := h.svc.ListTopicsBySection(c.Request().Context(), id)
	if err != nil {
		slog.Error("admin: list topics", "error", err)
		return c.String(http.StatusInternalServerError, "Внутренняя ошибка сервера")
	}
	return c.Render(http.StatusOK, "templates/admin/forum/topics_list.html", topicsListData{
		User:      auth.UserFromContext(c),
		CSRFToken: appMiddleware.CSRFToken(c),
		Section:   section,
		Topics:    topics,
	})
}

// NewTopic handles GET /admin/forum/sections/:id/topics/new.
func (h *ForumHandler) NewTopic(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return c.String(http.StatusBadRequest, "Неверный ID")
	}
	section, err := h.svc.GetSection(c.Request().Context(), id)
	if err != nil {
		slog.Error("admin: new topic get section", "error", err)
		return c.String(http.StatusInternalServerError, "Внутренняя ошибка сервера")
	}
	if section == nil {
		return c.String(http.StatusNotFound, "Раздел не найден")
	}
	return c.Render(http.StatusOK, "templates/admin/forum/topic_edit.html", topicEditData{
		User:      auth.UserFromContext(c),
		CSRFToken: appMiddleware.CSRFToken(c),
		Section:   section,
	})
}

// CreateTopic handles POST /admin/forum/sections/:id/topics.
func (h *ForumHandler) CreateTopic(c echo.Context) error {
	sectionID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return c.String(http.StatusBadRequest, "Неверный ID")
	}
	section, err := h.svc.GetSection(c.Request().Context(), sectionID)
	if err != nil {
		slog.Error("admin: create topic get section", "error", err)
		return c.String(http.StatusInternalServerError, "Внутренняя ошибка сервера")
	}
	if section == nil {
		return c.String(http.StatusNotFound, "Раздел не найден")
	}

	u := auth.UserFromContext(c)
	title := c.FormValue("title")
	_, err = h.svc.CreateTopic(c.Request().Context(), sectionID, u.ID, title)
	if err != nil {
		if errors.Is(err, forum.ErrEmptyTitle) || errors.Is(err, forum.ErrTitleTooLong) {
			return c.Render(http.StatusBadRequest, "templates/admin/forum/topic_edit.html", topicEditData{
				User:      u,
				CSRFToken: appMiddleware.CSRFToken(c),
				Section:   section,
				Error:     err.Error(),
			})
		}
		if errors.Is(err, forum.ErrSectionNotFound) {
			return c.String(http.StatusNotFound, "Раздел не найден")
		}
		slog.Error("admin: create topic", "error", err)
		return c.String(http.StatusInternalServerError, "Внутренняя ошибка сервера")
	}
	return c.Redirect(http.StatusSeeOther, "/admin/forum/sections/"+strconv.FormatInt(sectionID, 10)+"/topics")
}

// EditTopic handles GET /admin/forum/topics/:id/edit.
func (h *ForumHandler) EditTopic(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return c.String(http.StatusBadRequest, "Неверный ID")
	}
	topic, err := h.svc.GetTopic(c.Request().Context(), id)
	if err != nil {
		slog.Error("admin: edit topic get", "error", err)
		return c.String(http.StatusInternalServerError, "Внутренняя ошибка сервера")
	}
	if topic == nil {
		return c.String(http.StatusNotFound, "Тема не найдена")
	}
	section, err := h.svc.GetSection(c.Request().Context(), topic.SectionID)
	if err != nil {
		slog.Error("admin: edit topic get section", "error", err)
		return c.String(http.StatusInternalServerError, "Внутренняя ошибка сервера")
	}
	return c.Render(http.StatusOK, "templates/admin/forum/topic_edit.html", topicEditData{
		User:      auth.UserFromContext(c),
		CSRFToken: appMiddleware.CSRFToken(c),
		Section:   section,
		Topic:     topic,
	})
}

// UpdateTopic handles POST /admin/forum/topics/:id.
func (h *ForumHandler) UpdateTopic(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return c.String(http.StatusBadRequest, "Неверный ID")
	}
	topic, err := h.svc.GetTopic(c.Request().Context(), id)
	if err != nil {
		slog.Error("admin: update topic get", "error", err)
		return c.String(http.StatusInternalServerError, "Внутренняя ошибка сервера")
	}
	if topic == nil {
		return c.String(http.StatusNotFound, "Тема не найдена")
	}

	title := c.FormValue("title")
	if err := h.svc.UpdateTopic(c.Request().Context(), id, title); err != nil {
		if errors.Is(err, forum.ErrEmptyTitle) || errors.Is(err, forum.ErrTitleTooLong) {
			section, _ := h.svc.GetSection(c.Request().Context(), topic.SectionID)
			return c.Render(http.StatusBadRequest, "templates/admin/forum/topic_edit.html", topicEditData{
				User:      auth.UserFromContext(c),
				CSRFToken: appMiddleware.CSRFToken(c),
				Section:   section,
				Topic:     topic,
				Error:     err.Error(),
			})
		}
		slog.Error("admin: update topic", "error", err)
		return c.String(http.StatusInternalServerError, "Внутренняя ошибка сервера")
	}
	return c.Redirect(http.StatusSeeOther, "/admin/forum/sections/"+strconv.FormatInt(topic.SectionID, 10)+"/topics")
}

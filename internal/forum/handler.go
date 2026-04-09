package forum

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/Pegorino82/lfcru_forum/internal/auth"
	appMiddleware "github.com/Pegorino82/lfcru_forum/internal/middleware"
	"github.com/labstack/echo/v4"
)

type Handler struct {
	svc      *Service
	renderer echo.Renderer
}

func NewHandler(svc *Service, renderer echo.Renderer) *Handler {
	return &Handler{svc: svc, renderer: renderer}
}

// Index renders GET /forum — list of sections
func (h *Handler) Index(c echo.Context) error {
	ctx := c.Request().Context()

	sections, err := h.svc.ListSections(ctx)
	if err != nil {
		return c.String(http.StatusInternalServerError, "Internal error")
	}

	u := auth.UserFromContext(c)
	canManage := u != nil && (u.Role == "moderator" || u.Role == "admin")

	data := map[string]interface{}{
		"Sections":  sections,
		"CanManage": canManage,
		"CSRFToken": appMiddleware.CSRFToken(c),
	}

	isHTMX := c.Request().Header.Get("HX-Request") == "true"
	if isHTMX {
		return c.Render(http.StatusOK, "templates/forum/index.html#content", data)
	}
	return c.Render(http.StatusOK, "templates/forum/index.html", data)
}

// ShowSection renders GET /forum/sections/:id
func (h *Handler) ShowSection(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return c.String(http.StatusNotFound, "Not found")
	}

	ctx := c.Request().Context()
	section, topics, err := h.svc.GetSectionWithTopics(ctx, id)
	if err != nil {
		return c.String(http.StatusInternalServerError, "Internal error")
	}
	if section == nil {
		return c.String(http.StatusNotFound, "Not found")
	}

	u := auth.UserFromContext(c)
	canManage := u != nil && (u.Role == "moderator" || u.Role == "admin")

	data := map[string]interface{}{
		"Section":   section,
		"Topics":    topics,
		"CanManage": canManage,
		"CSRFToken": appMiddleware.CSRFToken(c),
	}

	isHTMX := c.Request().Header.Get("HX-Request") == "true"
	if isHTMX {
		return c.Render(http.StatusOK, "templates/forum/section.html#content", data)
	}
	return c.Render(http.StatusOK, "templates/forum/section.html", data)
}

// ShowTopic renders GET /forum/topics/:id
func (h *Handler) ShowTopic(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return c.String(http.StatusNotFound, "Not found")
	}

	ctx := c.Request().Context()
	topic, posts, err := h.svc.GetTopicWithPosts(ctx, id)
	if err != nil {
		return c.String(http.StatusInternalServerError, "Internal error")
	}
	if topic == nil {
		return c.String(http.StatusNotFound, "Not found")
	}

	// Fetch section for breadcrumbs
	section, err := h.svc.GetSection(ctx, topic.SectionID)
	if err != nil {
		return c.String(http.StatusInternalServerError, "Internal error")
	}

	u := auth.UserFromContext(c)
	canReply := u != nil

	data := map[string]interface{}{
		"Topic":     topic,
		"Section":   section,
		"Posts":     posts,
		"CanReply":  canReply,
		"CSRFToken": appMiddleware.CSRFToken(c),
	}

	isHTMX := c.Request().Header.Get("HX-Request") == "true"
	if isHTMX {
		return c.Render(http.StatusOK, "templates/forum/topic.html#content", data)
	}
	return c.Render(http.StatusOK, "templates/forum/topic.html", data)
}

// NewSection renders GET /forum/sections/new — form
func (h *Handler) NewSection(c echo.Context) error {
	data := map[string]interface{}{
		"CSRFToken": appMiddleware.CSRFToken(c),
	}

	isHTMX := c.Request().Header.Get("HX-Request") == "true"
	if isHTMX {
		return c.Render(http.StatusOK, "templates/forum/new_section.html#content", data)
	}
	return c.Render(http.StatusOK, "templates/forum/new_section.html", data)
}

// CreateSection handles POST /forum/sections
func (h *Handler) CreateSection(c echo.Context) error {
	title := c.FormValue("title")
	description := c.FormValue("description")
	sortOrderStr := c.FormValue("sort_order")

	sortOrder := 0
	if sortOrderStr != "" {
		so, err := strconv.Atoi(sortOrderStr)
		if err == nil {
			sortOrder = so
		}
	}

	ctx := c.Request().Context()
	_, err := h.svc.CreateSection(ctx, title, description, sortOrder)

	if err != nil {
		// Validation errors -> 422
		if errors.Is(err, ErrEmptyTitle) || errors.Is(err, ErrTitleTooLong) || errors.Is(err, ErrDescriptionTooLong) {
			errMsg := mapErrorMessage(err)
			data := map[string]interface{}{
				"FormError": errMsg,
				"FormContent": map[string]string{
					"title":       title,
					"description": description,
				},
				"CSRFToken": appMiddleware.CSRFToken(c),
			}
			return c.Render(http.StatusUnprocessableEntity, "templates/forum/new_section.html", data)
		}
		return c.String(http.StatusInternalServerError, "Internal error")
	}

	// Success -> redirect
	return c.Redirect(http.StatusSeeOther, "/forum")
}

// NewTopic renders GET /forum/sections/:id/topics/new
func (h *Handler) NewTopic(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return c.String(http.StatusNotFound, "Not found")
	}

	ctx := c.Request().Context()
	section, err := h.svc.GetSection(ctx, id)
	if err != nil {
		return c.String(http.StatusInternalServerError, "Internal error")
	}
	if section == nil {
		return c.String(http.StatusNotFound, "Not found")
	}

	data := map[string]interface{}{
		"Section":   section,
		"CSRFToken": appMiddleware.CSRFToken(c),
	}

	isHTMX := c.Request().Header.Get("HX-Request") == "true"
	if isHTMX {
		return c.Render(http.StatusOK, "templates/forum/new_topic.html#content", data)
	}
	return c.Render(http.StatusOK, "templates/forum/new_topic.html", data)
}

// CreateTopic handles POST /forum/sections/:id/topics
func (h *Handler) CreateTopic(c echo.Context) error {
	sectionID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return c.String(http.StatusNotFound, "Not found")
	}

	title := c.FormValue("title")

	u := auth.UserFromContext(c)
	ctx := c.Request().Context()

	topicID, err := h.svc.CreateTopic(ctx, sectionID, u.ID, title)

	if err != nil {
		// 404 if section not found
		if errors.Is(err, ErrSectionNotFound) {
			return c.String(http.StatusNotFound, "Not found")
		}
		// Validation errors -> 422
		if errors.Is(err, ErrEmptyTitle) || errors.Is(err, ErrTitleTooLong) {
			errMsg := mapErrorMessage(err)
			section, _ := h.svc.GetSection(ctx, sectionID)
			data := map[string]interface{}{
				"FormError": errMsg,
				"FormContent": map[string]string{
					"title": title,
				},
				"Section":   section,
				"CSRFToken": appMiddleware.CSRFToken(c),
			}
			return c.Render(http.StatusUnprocessableEntity, "templates/forum/new_topic.html", data)
		}
		return c.String(http.StatusInternalServerError, "Internal error")
	}

	// Success -> redirect
	return c.Redirect(http.StatusSeeOther, "/forum/topics/"+strconv.FormatInt(topicID, 10))
}

// CreatePost handles POST /forum/topics/:id/posts
func (h *Handler) CreatePost(c echo.Context) error {
	topicID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return c.String(http.StatusNotFound, "Not found")
	}

	content := c.FormValue("content")
	parentIDStr := c.FormValue("parent_id")

	var parentID *int64
	if parentIDStr != "" {
		pID, err := strconv.ParseInt(parentIDStr, 10, 64)
		if err == nil {
			parentID = &pID
		}
	}

	u := auth.UserFromContext(c)
	ctx := c.Request().Context()

	postID, err := h.svc.CreatePost(ctx, topicID, u.ID, parentID, content)

	if err != nil {
		// 404 if topic not found
		if errors.Is(err, ErrTopicNotFound) {
			return c.String(http.StatusNotFound, "Not found")
		}
		// Validation/logic errors -> 422
		if errors.Is(err, ErrEmptyContent) || errors.Is(err, ErrContentTooLong) ||
			errors.Is(err, ErrParentNotFound) || errors.Is(err, ErrReplyToReply) {
			errMsg := mapErrorMessage(err)

			isHTMX := c.Request().Header.Get("HX-Request") == "true"
			if isHTMX {
				// HTMX: return 422 with form + HX-Retarget
				c.Response().Header().Set("HX-Retarget", "#post-form")
				c.Response().Header().Set("HX-Reswap", "innerHTML")
				return c.String(http.StatusUnprocessableEntity, "Error: "+errMsg)
			}

			// Non-HTMX: return full page with error
			topic, posts, _ := h.svc.GetTopicWithPosts(ctx, topicID)
			section, _ := h.svc.GetSection(ctx, topic.SectionID)
			data := map[string]interface{}{
				"Topic":   topic,
				"Section": section,
				"Posts":   posts,
				"FormError": errMsg,
				"FormContent": map[string]string{
					"content":   content,
					"parent_id": parentIDStr,
				},
				"CanReply":  true,
				"CSRFToken": appMiddleware.CSRFToken(c),
			}
			return c.Render(http.StatusUnprocessableEntity, "templates/forum/topic.html", data)
		}
		return c.String(http.StatusInternalServerError, "Internal error")
	}

	// Success
	isHTMX := c.Request().Header.Get("HX-Request") == "true"
	if isHTMX {
		// HTMX: return updated posts list
		topic, posts, _ := h.svc.GetTopicWithPosts(ctx, topicID)
		data := map[string]interface{}{
			"Posts":     posts,
			"Topic":     topic,
			"CanReply":  true,
			"CSRFToken": appMiddleware.CSRFToken(c),
		}
		c.Response().Header().Set("HX-Trigger", "postAdded")
		return c.Render(http.StatusCreated, "templates/forum/topic.html#posts-list", data)
	}

	// Non-HTMX: redirect with anchor
	return c.Redirect(http.StatusSeeOther, "/forum/topics/"+strconv.FormatInt(topicID, 10)+"#post-"+strconv.FormatInt(postID, 10))
}

func mapErrorMessage(err error) string {
	switch {
	case errors.Is(err, ErrEmptyTitle):
		return "Заголовок не может быть пустым"
	case errors.Is(err, ErrTitleTooLong):
		return "Заголовок не может быть длиннее 255 символов"
	case errors.Is(err, ErrDescriptionTooLong):
		return "Описание не может быть длиннее 2000 символов"
	case errors.Is(err, ErrEmptyContent):
		return "Сообщение не может быть пустым"
	case errors.Is(err, ErrContentTooLong):
		return "Сообщение не может быть длиннее 20000 символов"
	case errors.Is(err, ErrParentNotFound):
		return "Цитируемое сообщение не найдено"
	case errors.Is(err, ErrReplyToReply):
		return "Нельзя цитировать ответ на ответ"
	default:
		return "Ошибка при обработке запроса"
	}
}
